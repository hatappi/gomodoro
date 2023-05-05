// Package cmd has versionCmd defined
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	version string
	commit  string
)

func newVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "show version",
		RunE: func(cmd *cobra.Command, args []string) error {
			short, err := cmd.Flags().GetBool("short")
			if err != nil {
				return err
			}

			if version == "" {
				version = "None"
			}

			if short {
				cmd.Println(version)
			} else {
				cmd.Printf("Version %s (git-%s)\n", version, commit)
			}

			return nil
		},
	}
	versionCmd.Flags().BoolP("short", "s", false, "show short version")

	return versionCmd
}
