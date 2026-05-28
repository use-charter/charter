package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := execute(); err != nil {
		cobra.CheckErr(err)
	}
}

func execute() error {
	cmd := newRootCommand()
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	return cmd.Execute()
}
