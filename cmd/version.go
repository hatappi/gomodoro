// Package cmd has versionCmd defined
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version string
	commit  string
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
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
			fmt.Println(version)
		} else {
			fmt.Printf("Version %s (git-%s)\n", version, commit)
		}

		return nil
	},
}

func init() {
	versionCmd.Flags().BoolP("short", "s", false, "show short version")
	rootCmd.AddCommand(versionCmd)
}
