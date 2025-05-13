package resolver

import (
	"context"

	"github.com/hatappi/go-kit/log"
	"github.com/hatappi/gomodoro/graph/model"
	"github.com/hatappi/gomodoro/internal/core/event"
)

var eventCategoryMap = map[model.EventCategory][]event.EventType{
	model.EventCategoryPomodoro: {
		event.PomodoroStarted, event.PomodoroPaused, event.PomodoroResumed,
		event.PomodoroCompleted, event.PomodoroStopped, event.PomodoroTick,
	},
	model.EventCategoryTask: {
		event.TaskCreated, event.TaskUpdated, event.TaskDeleted, event.TaskCompleted,
	},
}

var eventTypeMap = map[model.EventType]event.EventType{
	model.EventTypePomodoroStarted:   event.PomodoroStarted,
	model.EventTypePomodoroPaused:    event.PomodoroPaused,
	model.EventTypePomodoroResumed:   event.PomodoroResumed,
	model.EventTypePomodoroCompleted: event.PomodoroCompleted,
	model.EventTypePomodoroStopped:   event.PomodoroStopped,
	model.EventTypePomodoroTick:      event.PomodoroTick,
	model.EventTypeTaskCreated:       event.TaskCreated,
	model.EventTypeTaskUpdated:       event.TaskUpdated,
	model.EventTypeTaskDeleted:       event.TaskDeleted,
	model.EventTypeTaskCompleted:     event.TaskCompleted,
}

// EventReceived is the resolver for the eventReceived field.
func (r *subscriptionResolver) EventReceived(ctx context.Context, input model.EventReceivedInput) (<-chan *model.Event, error) {
	var eventTypes []event.EventType

	for _, cat := range input.EventCategory {
		if types, ok := eventCategoryMap[cat]; ok {
			eventTypes = append(eventTypes, types...)
		}
	}

	for _, t := range input.EventTypes {
		if typ, ok := eventTypeMap[t]; ok {
			eventTypes = append(eventTypes, typ)
		}
	}

	if len(eventTypes) == 0 {
		eventTypes = append(eventTypes, event.AllEventTypes...)
	}

	eventTypeStrs := make([]string, 0, len(eventTypes))
	for _, t := range eventTypes {
		eventTypeStrs = append(eventTypeStrs, string(t))
	}

	busCh, unsubscribe := r.EventBus.SubscribeChannel(eventTypeStrs)

	outCh := make(chan *model.Event)

	go func() {
		defer close(outCh)
		defer unsubscribe()

		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-busCh:
				log.FromContext(ctx).Info("event received", "event", e)
				if !ok {
					return
				}
				var ev *model.Event
				switch evt := e.(type) {
				case event.PomodoroEvent:
					payload := &model.EventPomodoroPayload{
						ID:            evt.ID,
						State:         toModelState(evt.State),
						RemainingTime: int32(evt.RemainingTime.Seconds()),
						ElapsedTime:   int32(evt.ElapsedTime.Seconds()),
						TaskID:        &evt.TaskID,
						Phase:         toModelPhase(evt.Phase),
						PhaseCount:    int32(evt.PhaseCount),
					}
					ev = &model.Event{
						EventCategory: model.EventCategoryPomodoro,
						EventType:     toModelEventType(evt.BaseEvent.Type),
						Payload:       payload,
					}
				case event.TaskEvent:
					payload := &model.EventTaskPayload{
						ID:        evt.ID,
						Title:     evt.Title,
						Completed: evt.Completed,
					}
					ev = &model.Event{
						EventCategory: model.EventCategoryTask,
						EventType:     toModelEventType(evt.BaseEvent.Type),
						Payload:       payload,
					}
				}
				if ev != nil {
					outCh <- ev
				}
			}
		}
	}()

	return outCh, nil
}

func toModelState(t event.PomodoroState) model.PomodoroState {
	switch t {
	case event.PomodoroStateActive:
		return model.PomodoroStateActive
	case event.PomodoroStatePaused:
		return model.PomodoroStatePaused
	case event.PomodoroStateFinished:
		return model.PomodoroStateFinished
	default:
		return ""
	}
}

func toModelPhase(t event.PomodoroPhase) model.PomodoroPhase {
	switch t {
	case event.PomodoroPhaseWork:
		return model.PomodoroPhaseWork
	case event.PomodoroPhaseShortBreak:
		return model.PomodoroPhaseShortBreak
	case event.PomodoroPhaseLongBreak:
		return model.PomodoroPhaseLongBreak
	default:
		return ""
	}
}

// toModelEventType converts event.EventType to model.EventType
func toModelEventType(t event.EventType) model.EventType {
	switch t {
	case event.PomodoroStarted:
		return model.EventTypePomodoroStarted
	case event.PomodoroPaused:
		return model.EventTypePomodoroPaused
	case event.PomodoroResumed:
		return model.EventTypePomodoroResumed
	case event.PomodoroCompleted:
		return model.EventTypePomodoroCompleted
	case event.PomodoroStopped:
		return model.EventTypePomodoroStopped
	case event.PomodoroTick:
		return model.EventTypePomodoroTick
	case event.TaskCreated:
		return model.EventTypeTaskCreated
	case event.TaskUpdated:
		return model.EventTypeTaskUpdated
	case event.TaskDeleted:
		return model.EventTypeTaskDeleted
	case event.TaskCompleted:
		return model.EventTypeTaskCompleted
	default:
		return ""
	}
}
