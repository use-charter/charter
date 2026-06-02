// Package tui implements the opt-in interactive browser for `charter doctor -i`.
//
// It is a Bubble Tea v2 (charm.land/bubbletea/v2) master-detail browser over a
// single doctor.Result: a findings list (charm.land/bubbles/v2/table) on the
// left, a scrollable detail pane (viewport) on the right, a brand/score header,
// and a help keybar (help + key) footer with a `/` search (textinput). It is
// reachable ONLY via `charter doctor -i` and only on a real TTY; the command
// layer gates that and the package never touches the non-interactive output
// contract.
//
// The model owns its filter, sort, and selection state directly so every
// transition is deterministic and unit testable by driving [Model.Update] with
// injected messages — no live screen is required for the tests. Colors come
// from internal/terminal (the same palette as the static renderer), degrading
// truecolor → 256 → 16 → mono.
package tui

import (
	"fmt"
	"io"

	tea "charm.land/bubbletea/v2"

	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/terminal"
)

// Run launches the interactive browser over result and blocks until the user
// quits (`q`/`ctrl+c`). scan re-runs the scan for the `r` rescan key; pass nil
// to disable it. out is the terminal the program renders to — the caller must
// have already confirmed it is a TTY (the command layer gates this and exits 2
// otherwise, so the program never runs headless).
//
// The browser uses the alternate screen and the v2 default synchronized output
// ("Mode 2026"); input is read from the controlling terminal automatically.
func Run(result doctor.Result, scan ScanFunc, caps terminal.Capabilities, pal terminal.Palette, out io.Writer) error {
	program := tea.NewProgram(New(result, scan, caps, pal), tea.WithOutput(out))
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("charter doctor interactive: %w", err)
	}
	return nil
}
