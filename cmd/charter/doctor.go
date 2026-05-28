package main

import (
	"errors"

	"github.com/spf13/cobra"
)

func newDoctorCommand() *cobra.Command {
	var path string
	var threshold int
	var quiet bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Scan a repository and compute a Charter score",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = path
			_ = threshold
			_ = quiet
			return errors.New("doctor runner not implemented")
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "explicit repository path")
	cmd.Flags().IntVar(&threshold, "threshold", 80, "minimum passing score")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "suppress non-failure output")

	return cmd
}
