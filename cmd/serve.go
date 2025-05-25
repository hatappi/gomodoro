// Package cmd implements CLI commands
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/hatappi/gomodoro/internal/api/server"
	"github.com/hatappi/gomodoro/internal/config"
)

// shutdownTimeout is the time to wait for the server to shutdown gracefully.
const shutdownTimeout = 5 * time.Second

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Gomodoro API server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			cfg, err := config.GetConfig()
			if err != nil {
				return fmt.Errorf("failed to get config: %w", err)
			}

			serverRunner := server.NewRunner(cfg)

			if err := serverRunner.Start(ctx); err != nil {
				return fmt.Errorf("failed to start server runner: %w", err)
			}

			ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()

			<-ctx.Done()

			shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()

			if err := serverRunner.Stop(shutdownCtx); err != nil {
				return fmt.Errorf("failed to stop server: %w", err)
			}

			return nil
		},
	}

	return cmd
}
