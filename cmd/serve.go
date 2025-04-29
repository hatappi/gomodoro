// Package cmd implements CLI commands
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hatappi/go-kit/log"
	"github.com/spf13/cobra"

	"github.com/hatappi/gomodoro/internal/api/server"
	"github.com/hatappi/gomodoro/internal/config"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Gomodoro API server",
		Long:  "Start the Gomodoro API server to enable CLI-API integration",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.GetConfig()
			if err != nil {
				return fmt.Errorf("failed to get config: %w", err)
			}

			logger := log.FromContext(ctx)
			serverRunner := server.NewRunner(cfg)
			ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()

			if err := serverRunner.Start(ctx); err != nil {
				return fmt.Errorf("failed to start server runner: %w", err)
			}
			logger.Info("API server started via runner")

			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := serverRunner.Stop(shutdownCtx); err != nil {
				return fmt.Errorf("error stopping server runner: %w", err)
			}

			logger.Info("API server stopped gracefully")
			return nil
		},
	}

	return cmd
}
