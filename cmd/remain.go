// Package cmd has startCmd defined
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hatappi/gomodoro/config"
	"github.com/hatappi/gomodoro/net/unix"
)

// remainCmd represents the remain command
var remainCmd = &cobra.Command{
	Use:   "remain",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.GetConfig()
		if err != nil {
			return err
		}

		c, err := unix.NewClient(config.UnixDomainScoketPath)
		if err != nil {
			return err
		}

		r, err := c.Get()
		if err != nil {
			return err
		}

		fmt.Printf("%s", r.GetRemain())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(remainCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// remainCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// remainCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
