package server

import (
	"context"
	"fmt"
	"time"

	"github.com/hatappi/go-kit/log"

	"github.com/hatappi/gomodoro/internal/pixela"
	"github.com/hatappi/gomodoro/internal/toggl"
)

// Option represents a function that configures the server.
type Option func(*Server)

// WithRecordToggl adds Toggl time tracking functionality.
func WithRecordToggl(togglClient *toggl.Client) Option {
	return func(a *Server) {
		a.completeFuncs = append(
			a.completeFuncs,
			func(ctx context.Context, taskName string, isWorkTime bool, elapsedTime time.Duration) error {
				if !isWorkTime {
					return nil
				}

				elapsedTimeSec := int(elapsedTime.Seconds())

				s := time.Now().Add(-time.Duration(elapsedTimeSec))

				if err := togglClient.PostTimeEntry(ctx, taskName, s, elapsedTimeSec); err != nil {
					return fmt.Errorf("failed to post time entry to Toggl: %w", err)
				}

				return nil
			},
		)
	}
}

// WithRecordPixela adds Pixela tracking functionality.
func WithRecordPixela(client *pixela.Client, userName, graphID string) Option {
	return func(a *Server) {
		a.completeFuncs = append(
			a.completeFuncs,
			func(ctx context.Context, _ string, isWorkTime bool, _ time.Duration) error {
				if !isWorkTime {
					return nil
				}

				if err := client.IncrementPixel(ctx, userName, graphID); err != nil {
					return fmt.Errorf("failed to increment pixel: %w", err)
				}

				return nil
			},
		)
	}
}

// WithCompletionLogging logs the completion of a pomodoro session.
func WithCompletionLogging() Option {
	return func(a *Server) {
		a.completeFuncs = append(
			a.completeFuncs,
			func(ctx context.Context, taskName string, isWorkTime bool, elapsedTime time.Duration) error {
				log.FromContext(ctx).Info(
					"Pomodoro completed",
					"taskName", taskName,
					"isWorkTime", isWorkTime,
					"elapsedTimeSec", elapsedTime.Seconds(),
				)

				return nil
			},
		)
	}
}
