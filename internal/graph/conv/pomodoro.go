package conv

import (
	"github.com/hatappi/gomodoro/internal/core"
	"github.com/hatappi/gomodoro/internal/graph/model"
)

// FromPomodoro converts a core.Pomodoro to a model.Pomodoro.
func FromPomodoro(pomodoro *core.Pomodoro) (*model.Pomodoro, error) {
	if pomodoro == nil {
		//nolint:nilnil
		return nil, nil
	}

	phase, err := convertEventPomodoroPhaseToModel(pomodoro.Phase)
	if err != nil {
		return nil, err
	}

	state, err := convertEventPomodoroStateToModel(pomodoro.State)
	if err != nil {
		return nil, err
	}

	return &model.Pomodoro{
		ID:        pomodoro.ID,
		State:     state,
		TaskID:    pomodoro.TaskID,
		StartTime: pomodoro.StartTime,
		Phase:     phase,

		PhaseCount:       pomodoro.PhaseCount,
		RemainingTimeSec: int(pomodoro.RemainingTime.Seconds()),
		ElapsedTimeSec:   int(pomodoro.ElapsedTime.Seconds()),
	}, nil
}
