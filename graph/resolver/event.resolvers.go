package resolver

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/hatappi/gomodoro/graph/model"
	"github.com/hatappi/gomodoro/internal/core/event"
)

// EventReceived is the resolver for the eventReceived field.
func (r *subscriptionResolver) EventReceived(
	ctx context.Context,
	input model.EventReceivedInput,
) (<-chan *model.Event, error) {
	eventTypes, err := getEventTypesFromInput(input)
	if err != nil {
		return nil, err
	}

	busCh, unsubscribe := r.EventBus.SubscribeChannel(eventTypes)
	outCh := make(chan *model.Event)

	go func() {
		defer close(outCh)
		defer unsubscribe()

		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-busCh:
				if !ok {
					return
				}

				switch evt := e.(type) {
				case event.PomodoroEvent:
					r.handlePomodoroEvent(ctx, evt, outCh)
				case event.TaskEvent:
					r.handleTaskEvent(ctx, evt, outCh)
				default:
					transport.AddSubscriptionError(ctx, gqlerror.Errorf("unknown event type: %T", evt))
				}
			}
		}
	}()

	return outCh, nil
}

// getEventTypesFromInput extracts event types from input.
func getEventTypesFromInput(input model.EventReceivedInput) ([]event.EventType, error) {
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

// handlePomodoroEvent processes a pomodoro event.
func (r *subscriptionResolver) handlePomodoroEvent(
	ctx context.Context,
	evt event.PomodoroEvent,
	outCh chan<- *model.Event,
) {
	state, err := convertEventPomodoroStateToModel(evt.State)
	if err != nil {
		transport.AddSubscriptionError(ctx, gqlerror.Errorf("failed to convert pomodoro state: %s", err))
		return
	}

	phase, err := convertEventPomodoroPhaseToModel(evt.Phase)
	if err != nil {
		transport.AddSubscriptionError(ctx, gqlerror.Errorf("failed to convert pomodoro phase: %s", err))
		return
	}

	eventType, err := convertEventTypeToModel(evt.BaseEvent.Type)
	if err != nil {
		transport.AddSubscriptionError(ctx, gqlerror.Errorf("failed to convert event type: %s", err))
		return
	}

	payload := &model.EventPomodoroPayload{
		ID:            evt.ID,
		State:         state,
		RemainingTime: int32(evt.RemainingTime.Seconds()),
		ElapsedTime:   int32(evt.ElapsedTime.Seconds()),
		TaskID:        &evt.TaskID,
		Phase:         phase,
		PhaseCount:    int32(evt.PhaseCount), // #nosec G115
	}

	outCh <- &model.Event{
		EventCategory: model.EventCategoryPomodoro,
		EventType:     eventType,
		Payload:       payload,
	}
}

// handleTaskEvent processes a task event.
func (r *subscriptionResolver) handleTaskEvent(ctx context.Context, evt event.TaskEvent, outCh chan<- *model.Event) {
	eventType, err := convertEventTypeToModel(evt.BaseEvent.Type)
	if err != nil {
		transport.AddSubscriptionError(ctx, gqlerror.Errorf("failed to convert event type: %s", err))
		return
	}

	payload := &model.EventTaskPayload{
		ID:        evt.ID,
		Title:     evt.Title,
		Completed: evt.Completed,
	}

	outCh <- &model.Event{
		EventCategory: model.EventCategoryTask,
		EventType:     eventType,
		Payload:       payload,
	}
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
			event.TaskCreated, event.TaskUpdated, event.TaskDeleted, event.TaskCompleted,
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
	case event.TaskCompleted:
		return model.EventTypeTaskCompleted, nil
	default:
		return "", fmt.Errorf("unknown event type: %s", t)
	}
}
