package main

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)


func main() {
	cmd := &cobra.Command{
		Use:   "restore [flags] RELEASE_NAME",
		Short: "restore last deployed release to original state",
		RunE:  run,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("RELEASE_NAME is required")
			}
			return nil
		},
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}

}

func run(cmd *cobra.Command, args []string) error {
	releaseName := args[0]
	if err := Restore(releaseName); err != nil {
		return err
	}
	return nil
}
