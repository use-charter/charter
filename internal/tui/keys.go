package tui

import "charm.land/bubbles/v2/key"

// keyMap is the Charter doctor browser's binding set. It implements help.KeyMap
// (ShortHelp/FullHelp) so the footer keybar renders directly from the bindings,
// keeping the help text and the Update handling in sync from one source.
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Switch   key.Binding
	Search   key.Binding
	Severity key.Binding
	Category key.Binding
	Muted    key.Binding
	Sort     key.Binding
	Copy     key.Binding
	Rescan   key.Binding
	Help     key.Binding
	Clear    key.Binding
	Quit     key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Switch: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch pane"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Severity: key.NewBinding(
			key.WithKeys("1", "2", "3", "4"),
			key.WithHelp("1-4", "severity"),
		),
		Category: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "category"),
		),
		Muted: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "suppressed"),
		),
		Sort: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "sort"),
		),
		Copy: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy path"),
		),
		Rescan: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rescan"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Clear: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp implements help.KeyMap: the compact footer keybar.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Switch, k.Search, k.Help, k.Quit}
}

// FullHelp implements help.KeyMap: the expanded help shown when `?` is toggled.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Switch},
		{k.Severity, k.Category, k.Muted, k.Sort},
		{k.Search, k.Clear, k.Copy, k.Rescan},
		{k.Help, k.Quit},
	}
}
