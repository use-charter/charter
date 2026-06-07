package main

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"go.use-charter.dev/charter/internal/doctor"
	renderhtml "go.use-charter.dev/charter/internal/render/html"
	renderjson "go.use-charter.dev/charter/internal/render/json"
	rendermarkdown "go.use-charter.dev/charter/internal/render/markdown"
)

// openInBrowser opens a written report with the OS default application. It is a
// package variable so tests can stub the opener instead of spawning a browser.
// Because it is shared mutable package state, any test that overrides it (e.g.
// TestReportOpen*) MUST NOT call t.Parallel(), or concurrent tests would race.
var openInBrowser = openWithOS

func newReportCommand() *cobra.Command {
	var path string
	var threshold int
	var format string
	var out string
	var open bool

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate a shareable Charter report (html, markdown, or json)",
		// report is a report GENERATOR, not a CI gate: it reuses the doctor scan
		// and renderers but always exits 0 on successful generation, regardless
		// of the score. `charter doctor` remains the gate that exits 1 below the
		// threshold. Render/write/usage errors exit 2.
		Long: "Generate a shareable Charter report from a repository scan.\n\n" +
			"report renders the same scan as `charter doctor` into a standalone artifact\n" +
			"(HTML by default, or Markdown/JSON) and writes it to a file. It is a report\n" +
			"generator, not a CI gate: it always exits 0 on successful generation regardless\n" +
			"of the score. Use `charter doctor` when you need an exit code that fails below\n" +
			"the threshold.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the renderer and default filename up front so an invalid
			// --format fails fast (exit 2) before the scan runs.
			defaultName := ""
			switch format {
			case "html":
				defaultName = "charter-report.html"
			case "markdown":
				defaultName = "charter-report.md"
			case "json":
				defaultName = "charter-report.json"
			default:
				return commandExitError{message: "invalid format: must be html, markdown, or json", exitCode: 2}
			}

			result, err := doctor.Run(path, threshold, cmd.Flags().Changed("threshold"))
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			// Reuse the existing renderers verbatim — no reimplementation.
			var data []byte
			switch format {
			case "html":
				data, err = renderhtml.Render(result)
			case "markdown":
				data, err = rendermarkdown.Render(result)
			case "json":
				data, err = renderjson.Render(result)
			}
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			// A report is always a file artifact: when --out is omitted, default
			// to the format-appropriate charter-report.* in the working dir.
			outPath := out
			if outPath == "" {
				outPath = defaultName
			}

			// outPath is always non-empty here (defaulted above), so emit takes
			// its file-writing branch: it normalizes to a single trailing
			// newline, writes the artifact, and prints `charter: wrote <path>`
			// to stderr — matching doctor's --out path.
			if err := emit(cmd, outPath, data); err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			// --open is best-effort: a failed opener is a note, not a failure,
			// since the artifact was already written successfully.
			if open {
				if openErr := openInBrowser(outPath); openErr != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "charter: could not open %s: %v\n", outPath, openErr)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "explicit repository path")
	cmd.Flags().IntVar(&threshold, "threshold", 80, "score threshold reflected in the report (report never gates on it)")
	cmd.Flags().StringVar(&format, "format", "html", "report format: html, markdown, or json")
	cmd.Flags().StringVar(&out, "out", "", "output file (defaults to charter-report.{html,md,json} for the chosen format)")
	cmd.Flags().BoolVar(&open, "open", false, "open the written report with the OS default application (best-effort)")

	return cmd
}

// openWithOS launches the platform default-application opener for path. It uses
// Start (non-blocking) so the CLI never waits on a GUI app, and treats a missing
// opener as an error the caller downgrades to a note. Offline: it only ever
// hands a local file path to the OS opener.
func openWithOS(path string) error {
	var name string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		name, args = "open", []string{path}
	case "windows":
		// explorer.exe opens the file without re-parsing the path through cmd.exe.
		name, args = "explorer.exe", []string{path}
	default:
		name, args = "xdg-open", []string{path}
	}
	// #nosec G204 -- name is a fixed per-OS opener and path is a local report
	// file the user just generated, not untrusted remote input.
	return exec.Command(name, args...).Start()
}
