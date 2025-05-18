// Package conv provides functions for converting between GraphQL and core types.
package conv

import (
	"fmt"
	"time"

	gqlgen "github.com/hatappi/gomodoro/internal/client/graphql/generated"
	"github.com/hatappi/gomodoro/internal/core"
	"github.com/hatappi/gomodoro/internal/core/event"
)

// ToCorePomodoro converts a GraphQL PomodoroDetails to a core Pomodoro.
func ToCorePomodoro(pomodoro gqlgen.PomodoroDetails) (*core.Pomodoro, error) {
	state, err := convertPomodoroStateToEvent(pomodoro.State)
	if err != nil {
		return nil, err
	}

	phase, err := convertPomodoroPhaseToEvent(pomodoro.Phase)
	if err != nil {
		return nil, err
	}

	return &core.Pomodoro{
		ID:            pomodoro.Id,
		State:         state,
		StartTime:     pomodoro.StartTime,
		TaskID:        pomodoro.TaskId,
		Phase:         phase,
		PhaseCount:    pomodoro.PhaseCount,
		RemainingTime: time.Duration(pomodoro.RemainingTimeSec),
		ElapsedTime:   time.Duration(pomodoro.ElapsedTimeSec),
	}, nil
}

func convertPomodoroStateToEvent(state gqlgen.PomodoroState) (event.PomodoroState, error) {
	switch state {
	case gqlgen.PomodoroStateActive:
		return event.PomodoroStateActive, nil
	case gqlgen.PomodoroStatePaused:
		return event.PomodoroStatePaused, nil
	case gqlgen.PomodoroStateFinished:
		return event.PomodoroStateFinished, nil
	default:
		return event.PomodoroState(""), fmt.Errorf("unknown pomodoro state: %s", state)
	}
}

func convertPomodoroPhaseToEvent(phase gqlgen.PomodoroPhase) (event.PomodoroPhase, error) {
	switch phase {
	case gqlgen.PomodoroPhaseWork:
		return event.PomodoroPhaseWork, nil
	case gqlgen.PomodoroPhaseShortBreak:
		return event.PomodoroPhaseShortBreak, nil
	case gqlgen.PomodoroPhaseLongBreak:
		return event.PomodoroPhaseLongBreak, nil
	default:
		return event.PomodoroPhase(""), fmt.Errorf("unknown pomodoro phase: %s", phase)
	}
}
