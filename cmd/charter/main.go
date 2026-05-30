package main

import (
	"errors"
	"fmt"
	"os"
)

type exitSignal interface {
	Silent() bool
	ExitCode() int
}

func main() {
	if err := execute(); err != nil {
		var signal exitSignal
		if ok := asExitSignal(err, &signal); ok {
			if !signal.Silent() && err.Error() != "" {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
			os.Exit(signal.ExitCode())
		}

		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func execute() error {
	cmd := newRootCommand()
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	return cmd.Execute()
}

func asExitSignal(err error, target *exitSignal) bool {
	return errors.As(err, target)
}
