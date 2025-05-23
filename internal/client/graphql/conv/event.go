package conv

import (
	"fmt"
	"time"

	gqlgen "github.com/hatappi/gomodoro/internal/client/graphql/generated"
	"github.com/hatappi/gomodoro/internal/core/event"
)

// ToEventInfo converts a GraphQL event to a core event.
//
//nolint:ireturn
func ToEventInfo(evt gqlgen.EventDetails) (event.EventInfo, error) {
	eventType, err := toEventType(evt.EventType)
	if err != nil {
		return nil, err
	}

	baseEvent := event.BaseEvent{
		Type:      eventType,
		Timestamp: time.Now(),
	}

	switch payload := evt.Payload.(type) {
	case *gqlgen.EventDetailsPayloadEventPomodoroPayload:
		return toPomodoroEvent(baseEvent, payload.EventPomodoroPayloadDetails)
	case *gqlgen.EventDetailsPayloadEventTaskPayload:
		return toTaskEvent(baseEvent, payload.EventTaskPayloadDetails), nil
	}

	return nil, fmt.Errorf("unknown event type: %s", evt.EventType)
}

func toTaskEvent(baseEvent event.BaseEvent, payload gqlgen.EventTaskPayloadDetails) event.TaskEvent {
	return event.TaskEvent{
		BaseEvent: baseEvent,
		ID:        payload.Id,
		Title:     payload.Title,
		Completed: payload.Completed,
	}
}

func toPomodoroEvent(
	baseEvent event.BaseEvent,
	payload gqlgen.EventPomodoroPayloadDetails,
) (event.PomodoroEvent, error) {
	state, err := convertPomodoroStateToEvent(payload.State)
	if err != nil {
		return event.PomodoroEvent{}, err
	}

	phase, err := convertPomodoroPhaseToEvent(payload.Phase)
	if err != nil {
		return event.PomodoroEvent{}, err
	}

	return event.PomodoroEvent{
		BaseEvent:     baseEvent,
		ID:            payload.Id,
		State:         state,
		RemainingTime: time.Duration(payload.RemainingTime) * time.Second,
		ElapsedTime:   time.Duration(payload.ElapsedTime) * time.Second,
		TaskID:        payload.TaskId,
		Phase:         phase,
		PhaseCount:    payload.PhaseCount,
	}, nil
}

func toEventType(eventType gqlgen.EventType) (event.EventType, error) {
	switch eventType {
	case gqlgen.EventTypePomodoroStarted:
		return event.PomodoroStarted, nil
	case gqlgen.EventTypePomodoroPaused:
		return event.PomodoroPaused, nil
	case gqlgen.EventTypePomodoroResumed:
		return event.PomodoroResumed, nil
	case gqlgen.EventTypePomodoroCompleted:
		return event.PomodoroCompleted, nil
	case gqlgen.EventTypePomodoroStopped:
		return event.PomodoroStopped, nil
	case gqlgen.EventTypePomodoroTick:
		return event.PomodoroTick, nil
	case gqlgen.EventTypeTaskCreated:
		return event.TaskCreated, nil
	case gqlgen.EventTypeTaskUpdated:
		return event.TaskUpdated, nil
	case gqlgen.EventTypeTaskDeleted:
		return event.TaskDeleted, nil
	case gqlgen.EventTypeTaskCompleted:
		return event.TaskCompleted, nil
	default:
		return event.EventType(""), fmt.Errorf("unknown event type: %s", eventType)
	}
}
