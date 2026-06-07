package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/terminal"
)

// focusArea is which pane currently receives navigation keys.
type focusArea int

const (
	focusList focusArea = iota
	focusDetail
)

// ScanFunc re-runs the scan for an in-place rescan (the `r` key). It is injected
// so the model carries no dependency on the command layer and is fully testable
// with a deterministic stub.
type ScanFunc func() (doctor.Result, error)

// rescanDoneMsg delivers the result of an in-place rescan back to Update.
type rescanDoneMsg struct {
	result doctor.Result
	err    error
}

// Model is the Bubble Tea v2 master-detail browser over a single doctor.Result.
// It owns its filter/sort/selection state directly (rather than delegating to a
// component's internal filter) so every transition is deterministic and unit
// testable by driving Update with injected messages.
//
// All methods take a value receiver and, when they mutate, return the updated
// Model (the idiomatic MVU shape, matching Update's reducer signature). Callers
// reassign the result (m = m.refresh()); there are no pointer-receiver methods,
// so a transition can never mutate state in place and be silently dropped.
type Model struct {
	// Inputs / config.
	caps terminal.Capabilities
	pal  terminal.Palette
	scan ScanFunc
	th   theme

	// Data.
	result     doctor.Result
	items      []item
	filtered   []item
	categories []string
	selected   int

	// Filter / sort state.
	sevFilter findings.Severity
	catFilter string
	showMuted bool
	query     string
	sort      sortMode

	// Interaction state.
	focus     focusArea
	searching bool
	showHelp  bool
	status    string
	quitting  bool
	ready     bool

	// Layout.
	width  int
	height int

	// Sub-components.
	table  table.Model
	detail viewport.Model
	search textinput.Model
	help   help.Model
	keys   keyMap
}

// New builds a browser model over result. scan re-runs the scan for the `r`
// rescan key; pass nil to disable rescanning. caps/pal carry the resolved
// terminal capabilities and Charter palette so the view styles to the detected
// color tier. The model starts ready with sane default dimensions so View and
// the unit tests work before the first WindowSizeMsg arrives.
func New(result doctor.Result, scan ScanFunc, caps terminal.Capabilities, pal terminal.Palette) Model {
	m := Model{
		caps:   caps,
		pal:    pal,
		scan:   scan,
		th:     newTheme(caps, pal),
		result: result,
		keys:   defaultKeyMap(),
		focus:  focusList,
		width:  defaultWidth,
		height: defaultHeight,
		ready:  true,
	}
	m.items = buildItems(result)
	m.categories = uniqueCategories(m.items)

	tbl := table.New(table.WithFocused(true))
	tbl.SetStyles(m.tableStyles())
	m.table = tbl

	vp := viewport.New()
	vp.SoftWrap = true
	m.detail = vp

	ti := textinput.New()
	ti.Prompt = "/"
	ti.Placeholder = "search findings"
	m.search = ti

	h := help.New()
	h.Styles = help.DefaultStyles(m.pal.DarkBackground())
	m.help = h

	m = m.layout()
	m = m.refresh()
	return m
}

// Init implements tea.Model. The browser is static (window size is delivered
// automatically), so no startup command is required.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model: the single reducer that maps every message to a
// new model and an optional command.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		m = m.layout()
		return m, nil
	case rescanDoneMsg:
		return m.applyRescan(msg), nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey routes a key press. Search mode captures most keys for the text
// input; otherwise the binding set drives filters, actions, focus, and nav.
func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.searching {
		return m.handleSearchKey(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		m.quitting = true
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
		m.help.ShowAll = m.showHelp
		m = m.layout()
		return m, nil
	case key.Matches(msg, m.keys.Search):
		m.searching = true
		m.focus = focusList
		m.status = ""
		cmd := m.search.Focus()
		return m, cmd
	case key.Matches(msg, m.keys.Switch):
		m = m.toggleFocus()
		return m, nil
	case key.Matches(msg, m.keys.Rescan):
		if m.scan == nil {
			m.status = "rescan unavailable"
			return m, nil
		}
		m.status = "rescanning…"
		return m, m.rescanCmd()
	case key.Matches(msg, m.keys.Copy):
		return m.copySelection()
	case key.Matches(msg, m.keys.Severity):
		m = m.applySeverityKey(msg.String())
		return m, nil
	case key.Matches(msg, m.keys.Category):
		m = m.cycleCategory()
		m.selected = 0
		m = m.refresh()
		return m, nil
	case key.Matches(msg, m.keys.Muted):
		m.showMuted = !m.showMuted
		m.selected = 0
		m = m.refresh()
		return m, nil
	case key.Matches(msg, m.keys.Sort):
		if m.sort == sortBySeverity {
			m.sort = sortByCategory
		} else {
			m.sort = sortBySeverity
		}
		m = m.refresh()
		return m, nil
	case key.Matches(msg, m.keys.Clear):
		m = m.clearFilters()
		return m, nil
	default:
		return m.handleNav(msg)
	}
}

// handleNav forwards a navigation key to the focused pane: the detail viewport
// scrolls, or the findings table moves its cursor (after which the detail pane
// is re-synced to the newly selected finding).
func (m Model) handleNav(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.focus == focusDetail {
		m.detail, cmd = m.detail.Update(msg)
		return m, cmd
	}
	m.table, cmd = m.table.Update(msg)
	if m.table.Cursor() != m.selected {
		m = m.syncDetail()
	}
	return m, cmd
}

// handleSearchKey handles keys while the `/` search input is focused. Enter
// commits the query (keeping the input's text as the active filter), esc
// cancels and clears it, and everything else is fed to the text input with the
// list re-filtered live as the user types.
func (m Model) handleSearchKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Clear):
		m.searching = false
		m.search.Blur()
		m.search.Reset()
		m.query = ""
		m.selected = 0
		m = m.refresh()
		return m, nil
	case msg.Code == tea.KeyEnter:
		m.searching = false
		m.search.Blur()
		m.query = m.search.Value()
		m.selected = 0
		m = m.refresh()
		return m, nil
	default:
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		m.query = m.search.Value()
		m.selected = 0
		m = m.refresh()
		return m, cmd
	}
}

// toggleFocus flips between the list and the detail pane, focusing/blurring the
// table so it only consumes navigation keys when it owns focus.
func (m Model) toggleFocus() Model {
	if m.focus == focusList {
		m.focus = focusDetail
		m.table.Blur()
	} else {
		m.focus = focusList
		m.table.Focus()
	}
	return m
}

// applySeverityKey sets (or toggles off) the severity filter from a 1-4 key.
func (m Model) applySeverityKey(s string) Model {
	sev := severityForDigit(s)
	if sev == "" {
		return m
	}
	if m.sevFilter == sev {
		m.sevFilter = "" // pressing the active severity again clears it
	} else {
		m.sevFilter = sev
	}
	m.selected = 0
	return m.refresh()
}

// cycleCategory advances the category filter through the sorted category set,
// wrapping from the last category back to "all".
func (m Model) cycleCategory() Model {
	if len(m.categories) == 0 {
		return m
	}
	if m.catFilter == "" {
		m.catFilter = m.categories[0]
		return m
	}
	for i, c := range m.categories {
		if c == m.catFilter {
			if i+1 >= len(m.categories) {
				m.catFilter = ""
			} else {
				m.catFilter = m.categories[i+1]
			}
			return m
		}
	}
	m.catFilter = ""
	return m
}

// clearFilters resets every filter and the search query to the default view.
func (m Model) clearFilters() Model {
	m.sevFilter = ""
	m.catFilter = ""
	m.showMuted = false
	m.query = ""
	m.search.Reset()
	m.selected = 0
	m.status = "filters cleared"
	return m.refresh()
}

// copySelection copies the selected finding's first path:line to the system
// clipboard via OSC 52. Terminals without OSC 52 support simply ignore the
// sequence, so the action degrades gracefully.
func (m Model) copySelection() (tea.Model, tea.Cmd) {
	if m.selected < 0 || m.selected >= len(m.filtered) {
		m.status = "nothing to copy"
		return m, nil
	}
	loc := firstLocation(m.filtered[m.selected].finding)
	if loc == "" {
		m.status = "no path:line to copy"
		return m, nil
	}
	m.status = "copied " + loc
	return m, tea.SetClipboard(loc)
}

// rescanCmd returns a command that re-runs the injected scan off the event loop
// and reports the outcome via rescanDoneMsg.
func (m Model) rescanCmd() tea.Cmd {
	scan := m.scan
	return func() tea.Msg {
		result, err := scan()
		return rescanDoneMsg{result: result, err: err}
	}
}

// applyRescan rebuilds the model from a completed rescan, preserving the active
// filters but resetting the selection to the top of the new list.
func (m Model) applyRescan(msg rescanDoneMsg) Model {
	if msg.err != nil {
		m.status = "rescan failed: " + msg.err.Error()
		return m
	}
	m.result = msg.result
	m.items = buildItems(msg.result)
	m.categories = uniqueCategories(m.items)
	m.selected = 0
	m.status = "rescanned"
	return m.refresh()
}

// refresh recomputes the filtered/sorted item set, rebuilds the table rows,
// clamps the selection, and re-renders the detail pane. It is the single place
// that derives view state from filter state, so callers only mutate filter
// fields and call refresh.
func (m Model) refresh() Model {
	m.filtered = filterItems(m.items, m.sevFilter, m.catFilter, m.showMuted, m.query)
	sortItems(m.filtered, m.sort)

	rows := make([]table.Row, len(m.filtered))
	for i, it := range m.filtered {
		rows[i] = m.rowFor(it)
	}
	m.table.SetRows(rows)

	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
	m.table.SetCursor(m.selected)
	return m.syncDetail()
}

// syncDetail aligns the selected index with the table cursor and refreshes the
// detail pane content for the current selection (or an empty-state message when
// nothing matches the filters).
func (m Model) syncDetail() Model {
	m.selected = m.table.Cursor()
	if m.selected < 0 || m.selected >= len(m.filtered) {
		m.detail.SetContent(m.renderEmptyDetail())
		return m
	}
	m.detail.SetContent(m.renderDetail(m.filtered[m.selected]))
	m.detail.GotoTop()
	return m
}

// selectedItem returns the currently selected item and whether one exists. It
// is exported to the package (and exercised by tests) as the single accessor
// for the active selection.
func (m Model) selectedItem() (item, bool) {
	if m.selected < 0 || m.selected >= len(m.filtered) {
		return item{}, false
	}
	return m.filtered[m.selected], true
}
