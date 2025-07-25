package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.73

import (
	"context"
	"fmt"
	"time"

	"github.com/hatappi/gomodoro/internal/graph/conv"
	"github.com/hatappi/gomodoro/internal/graph/model"
)

// StartPomodoro is the resolver for the startPomodoro field.
func (r *mutationResolver) StartPomodoro(ctx context.Context, input model.StartPomodoroInput) (*model.Pomodoro, error) {
	pomodoro, err := r.PomodoroService.Start(
		ctx,
		time.Duration(input.WorkDurationSec)*time.Second,
		time.Duration(input.BreakDurationSec)*time.Second,
		time.Duration(input.LongBreakDurationSec)*time.Second,
		input.TaskID,
	)
	if err != nil {
		return nil, err
	}

	return conv.FromPomodoro(pomodoro)
}

// PausePomodoro is the resolver for the pausePomodoro field.
func (r *mutationResolver) PausePomodoro(ctx context.Context) (*model.Pomodoro, error) {
	activePomodoro, err := r.PomodoroService.ActivePomodoro()
	if err != nil {
		return nil, err
	}

	pomodoro, err := r.PomodoroService.Pause(ctx, activePomodoro.ID)
	if err != nil {
		return nil, err
	}

	return conv.FromPomodoro(pomodoro)
}

// ResumePomodoro is the resolver for the resumePomodoro field.
func (r *mutationResolver) ResumePomodoro(ctx context.Context) (*model.Pomodoro, error) {
	activePomodoro, err := r.PomodoroService.ActivePomodoro()
	if err != nil {
		return nil, err
	}

	pomodoro, err := r.PomodoroService.Resume(ctx, activePomodoro.ID)
	if err != nil {
		return nil, err
	}

	return conv.FromPomodoro(pomodoro)
}

// StopPomodoro is the resolver for the stopPomodoro field.
func (r *mutationResolver) StopPomodoro(ctx context.Context) (*model.Pomodoro, error) {
	activePomodoro, err := r.PomodoroService.ActivePomodoro()
	if err != nil {
		return nil, err
	}

	if err := r.PomodoroService.Stop(ctx, activePomodoro.ID); err != nil {
		return nil, err
	}

	return conv.FromPomodoro(activePomodoro)
}

// CurrentPomodoro is the resolver for the currentPomodoro field.
func (r *queryResolver) CurrentPomodoro(ctx context.Context) (*model.Pomodoro, error) {
	pomodoro, err := r.PomodoroService.LatestPomodoro()
	if err != nil {
		return nil, err
	}

	if pomodoro == nil {
		return nil, fmt.Errorf("no current pomodoro")
	}

	return conv.FromPomodoro(pomodoro)
}
