// Package conv provides functions for converting between different event types.
package conv

import (
	"fmt"

	"github.com/hatappi/gomodoro/internal/core/event"
	"github.com/hatappi/gomodoro/internal/graph/model"
)

// GetEventTypesFromInput extracts event types from input.
func GetEventTypesFromInput(input model.EventReceivedInput) ([]event.EventType, error) {
	var eventTypes []event.EventType

	for _, cat := range input.EventCategory {
		types, err := convertModelEventCategoryToEventTypes(cat)
		if err != nil {
			return nil, fmt.Errorf("failed to convert event category: %w", err)
		}
		eventTypes = append(eventTypes, types...)
	}

	if len(eventTypes) == 0 {
		eventTypes = append(eventTypes, event.AllEventTypes...)
	}

	return eventTypes, nil
}

// ConvertPomodoroEventToModelEvent processes a pomodoro event.
func ConvertPomodoroEventToModelEvent(evt event.PomodoroEvent) (*model.Event, error) {
	state, err := convertEventPomodoroStateToModel(evt.State)
	if err != nil {
		return nil, fmt.Errorf("failed to convert pomodoro state: %w", err)
	}

	phase, err := convertEventPomodoroPhaseToModel(evt.Phase)
	if err != nil {
		return nil, fmt.Errorf("failed to convert pomodoro phase: %w", err)
	}

	eventType, err := convertEventTypeToModel(evt.BaseEvent.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to convert event type: %w", err)
	}

	payload := &model.EventPomodoroPayload{
		ID:            evt.ID,
		State:         state,
		RemainingTime: evt.RemainingTime,
		ElapsedTime:   evt.ElapsedTime,
		TaskID:        &evt.TaskID,
		Phase:         phase,
		PhaseCount:    evt.PhaseCount,
	}

	return &model.Event{
		EventCategory: model.EventCategoryPomodoro,
		EventType:     eventType,
		Payload:       payload,
	}, nil
}

// ConvertTaskEventToModelEvent processes a task event.
func ConvertTaskEventToModelEvent(evt event.TaskEvent) (*model.Event, error) {
	eventType, err := convertEventTypeToModel(evt.BaseEvent.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to convert event type: %w", err)
	}

	payload := &model.EventTaskPayload{
		ID:    evt.ID,
		Title: evt.Title,
	}

	return &model.Event{
		EventCategory: model.EventCategoryTask,
		EventType:     eventType,
		Payload:       payload,
	}, nil
}

func convertModelEventCategoryToEventTypes(mcat model.EventCategory) ([]event.EventType, error) {
	switch mcat {
	case model.EventCategoryPomodoro:
		return []event.EventType{
			event.PomodoroStarted, event.PomodoroPaused, event.PomodoroResumed,
			event.PomodoroCompleted, event.PomodoroStopped, event.PomodoroTick,
		}, nil
	case model.EventCategoryTask:
		return []event.EventType{
			event.TaskCreated, event.TaskUpdated, event.TaskDeleted,
		}, nil
	default:
		return nil, fmt.Errorf("unknown event category: %s", mcat)
	}
}

func convertEventPomodoroStateToModel(t event.PomodoroState) (model.PomodoroState, error) {
	switch t {
	case event.PomodoroStateActive:
		return model.PomodoroStateActive, nil
	case event.PomodoroStatePaused:
		return model.PomodoroStatePaused, nil
	case event.PomodoroStateFinished:
		return model.PomodoroStateFinished, nil
	default:
		return "", fmt.Errorf("unknown pomodoro state: %s", t)
	}
}

func convertEventPomodoroPhaseToModel(t event.PomodoroPhase) (model.PomodoroPhase, error) {
	switch t {
	case event.PomodoroPhaseWork:
		return model.PomodoroPhaseWork, nil
	case event.PomodoroPhaseShortBreak:
		return model.PomodoroPhaseShortBreak, nil
	case event.PomodoroPhaseLongBreak:
		return model.PomodoroPhaseLongBreak, nil
	default:
		return "", fmt.Errorf("unknown pomodoro phase: %s", t)
	}
}

func convertEventTypeToModel(t event.EventType) (model.EventType, error) {
	switch t {
	case event.PomodoroStarted:
		return model.EventTypePomodoroStarted, nil
	case event.PomodoroPaused:
		return model.EventTypePomodoroPaused, nil
	case event.PomodoroResumed:
		return model.EventTypePomodoroResumed, nil
	case event.PomodoroCompleted:
		return model.EventTypePomodoroCompleted, nil
	case event.PomodoroStopped:
		return model.EventTypePomodoroStopped, nil
	case event.PomodoroTick:
		return model.EventTypePomodoroTick, nil
	case event.TaskCreated:
		return model.EventTypeTaskCreated, nil
	case event.TaskUpdated:
		return model.EventTypeTaskUpdated, nil
	case event.TaskDeleted:
		return model.EventTypeTaskDeleted, nil
	default:
		return "", fmt.Errorf("unknown event type: %s", t)
	}
}
