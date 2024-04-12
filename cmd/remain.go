// Package cmd has startCmd defined
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hatappi/gomodoro/internal/config"
	"github.com/hatappi/gomodoro/internal/net/unix"
)

func newRemainCmd() *cobra.Command {
	remainCmd := &cobra.Command{
		Use:   "remain",
		Short: "get remain time",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			ie, e := cmd.Flags().GetBool("ignore-error")
			if e != nil {
				return e
			}

			err := func() error {
				config, err := config.GetConfig()
				if err != nil {
					return err
				}

				c, err := unix.NewClient(config.UnixDomainScoketPath)
				if err != nil {
					return err
				}

				r, err := c.Get(ctx)
				if err != nil {
					return err
				}

				fmt.Printf("%s", r.GetRemain())
				return nil
			}()
			if err != nil {
				if !ie {
					return err
				}
				fmt.Printf("--:--")
			}

			return nil
		},
	}

	remainCmd.Flags().BoolP("ignore-error", "i", false, "ignore error")

	return remainCmd
}
