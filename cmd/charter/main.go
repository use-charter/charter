package main

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/charmbracelet/fang"

	"go.use-charter.dev/charter/internal/version"
)

type exitSignal interface {
	Silent() bool
	ExitCode() int
}

func main() {
	if err := execute(); err != nil {
		var signal exitSignal
		if asExitSignal(err, &signal) {
			os.Exit(signal.ExitCode())
		}
		os.Exit(2)
	}
}

// execute wraps Charter's Cobra root in fang for styled help/usage/error output
// and a `--version` flag, then runs it.
//
// fang.Execute (v1.0.0) never calls os.Exit: it runs the command, routes any
// resulting error through the configured ErrorHandler (which owns *printing*),
// and returns that same error. That split lets main() stay the single owner of
// the process exit code, so Charter's exit-code + silent-error contract is
// preserved exactly:
//   - silentAwareErrorHandler suppresses output for silent exitSignal errors
//     (the doctor below-threshold FAIL, which already printed its report) and
//     styles every other error via fang's default handler;
//   - main() inspects the returned error and exits with ExitCode() for an
//     exitSignal, else 2 — identical routing to the pre-fang implementation.
func execute() error {
	cmd := newRootCommand()
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	return fang.Execute(
		context.Background(),
		cmd,
		fang.WithVersion(version.Version()),
		fang.WithCommit(version.Commit()),
		fang.WithErrorHandler(silentAwareErrorHandler),
		// Man-page generation would add a net-new (hidden) `man` subcommand;
		// skip it to preserve the existing subcommand set. Completions stay on
		// (fang's default), matching the cobra default command this CLI already
		// exposed, so no existing behavior changes.
		fang.WithoutManpage(),
	)
}

// silentAwareErrorHandler preserves Charter's silent-error contract on top of
// fang's styling. fang invokes the ErrorHandler for EVERY non-nil error before
// returning it, so a silent exitSignal must be swallowed here to guarantee no
// error box is printed; main() still exits with the signal's code. Every other
// error is rendered through fang's default styled handler, which downgrades to
// plain text when stderr is not a TTY or NO_COLOR is set.
func silentAwareErrorHandler(w io.Writer, styles fang.Styles, err error) {
	var signal exitSignal
	// Mirror the pre-fang print guard (`!signal.Silent() && err.Error() != ""`)
	// exactly: suppress output for a silent exitSignal, or for an exitSignal
	// with an empty message. The empty-message half is defensive — every
	// commandExitError carries a message today — but kept for parity.
	if errors.As(err, &signal) && (signal.Silent() || err.Error() == "") {
		return
	}
	fang.DefaultErrorHandler(w, styles, err)
}

func asExitSignal(err error, target *exitSignal) bool {
	return errors.As(err, target)
}
