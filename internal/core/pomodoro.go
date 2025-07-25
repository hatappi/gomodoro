// Package core provides the core business logic for the application
package core

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/hatappi/go-kit/log"

	"github.com/hatappi/gomodoro/internal/core/event"
	"github.com/hatappi/gomodoro/internal/storage"
)

// Pomodoro represents a pomodoro session with its current state.
type Pomodoro struct {
	ID            string              `json:"id"`
	State         event.PomodoroState `json:"state"`
	StartTime     time.Time           `json:"start_time"`
	WorkDuration  time.Duration       `json:"work_duration"`
	BreakDuration time.Duration       `json:"break_duration"`
	RemainingTime time.Duration       `json:"remaining_time"`
	ElapsedTime   time.Duration       `json:"elapsed_time"`
	Phase         event.PomodoroPhase `json:"phase"`
	PhaseCount    int                 `json:"phase_count"`
	TaskID        string              `json:"task_id,omitempty"`
}

// PomodoroService provides operations for managing pomodoro sessions.
type PomodoroService struct {
	storage  storage.PomodoroStorage
	eventBus event.EventBus
	ticker   *time.Ticker
	stopChan chan struct{}
}

// NewPomodoroService creates a new pomodoro service instance.
func NewPomodoroService(storage storage.PomodoroStorage, eventBus event.EventBus) *PomodoroService {
	return &PomodoroService{
		storage:  storage,
		eventBus: eventBus,
		stopChan: make(chan struct{}),
	}
}

// Start begins a new pomodoro session.
func (s *PomodoroService) Start(
	ctx context.Context,
	workDuration,
	breakDuration time.Duration,
	longBreakDuration time.Duration,
	taskID string,
) (*Pomodoro, error) {
	latestPomodoro, err := s.LatestPomodoro()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest pomodoro: %w", err)
	}

	if latestPomodoro != nil && latestPomodoro.State == event.PomodoroStateActive {
		return nil, fmt.Errorf("active pomodoro session already exists")
	}

	phase, duration, phaseCount := s.determinePhaseAndDuration(
		latestPomodoro,
		workDuration,
		breakDuration,
		longBreakDuration,
	)

	pomodoro := &storage.Pomodoro{
		ID:                uuid.New().String(),
		State:             storage.PomodoroStateActive,
		StartTime:         time.Now(),
		WorkDuration:      workDuration,
		BreakDuration:     breakDuration,
		LongBreakDuration: longBreakDuration,
		RemainingTime:     duration,
		Phase:             phase,
		PhaseCount:        phaseCount,
		TaskID:            taskID,
	}

	if err := s.storage.SavePomodoro(pomodoro); err != nil {
		return nil, fmt.Errorf("failed to save pomodoro: %w", err)
	}

	s.publishPomodoroEvent(event.PomodoroStarted, pomodoro)

	s.startTimer(ctx, pomodoro.ID, duration)

	return s.storagePomodoroToCore(pomodoro), nil
}

// Pause pauses an active pomodoro session.
func (s *PomodoroService) Pause(_ context.Context, id string) (*Pomodoro, error) {
	s.stopTimer()

	activePomodoro, err := s.storage.GetActivePomodoro()
	if err != nil {
		return nil, fmt.Errorf("failed to get active pomodoro: %w", err)
	}

	if activePomodoro == nil {
		return nil, fmt.Errorf("no active pomodoro found")
	}

	pomodoro, err := s.storage.UpdatePomodoroState(
		id,
		storage.PomodoroStatePaused,
		int(activePomodoro.RemainingTime.Seconds()),
		int(activePomodoro.ElapsedTime.Seconds()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update pomodoro state: %w", err)
	}

	s.publishPomodoroEvent(event.PomodoroPaused, pomodoro)

	return s.storagePomodoroToCore(pomodoro), nil
}

// Resume resumes a paused pomodoro session.
func (s *PomodoroService) Resume(ctx context.Context, id string) (*Pomodoro, error) {
	activePomodoro, err := s.storage.GetActivePomodoro()
	if err != nil {
		return nil, fmt.Errorf("failed to get active pomodoro: %w", err)
	}

	if activePomodoro == nil {
		return nil, fmt.Errorf("no active pomodoro found")
	}

	if activePomodoro.State != storage.PomodoroStatePaused {
		return nil, fmt.Errorf("pomodoro is not paused")
	}

	pomodoro, err := s.storage.UpdatePomodoroState(
		id,
		storage.PomodoroStateActive,
		int(activePomodoro.RemainingTime.Seconds()),
		int(activePomodoro.ElapsedTime.Seconds()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update pomodoro state: %w", err)
	}

	s.publishPomodoroEvent(event.PomodoroResumed, pomodoro)

	s.startTimer(ctx, id, pomodoro.RemainingTime)

	return s.storagePomodoroToCore(pomodoro), nil
}

// Stop stops the current pomodoro session.
func (s *PomodoroService) Stop(_ context.Context, id string) error {
	s.stopTimer()

	activePomodoro, err := s.storage.GetActivePomodoro()
	if err != nil {
		return fmt.Errorf("failed to get active pomodoro: %w", err)
	}

	if activePomodoro == nil {
		return fmt.Errorf("no active pomodoro found")
	}

	pomodoro, err := s.storage.UpdatePomodoroState(
		id,
		storage.PomodoroStateFinished,
		0,
		int(activePomodoro.ElapsedTime.Seconds()),
	)
	if err != nil {
		return fmt.Errorf("failed to update pomodoro state: %w", err)
	}

	s.publishPomodoroEvent(event.PomodoroStopped, pomodoro)

	return nil
}

// Delete deletes a pomodoro session by ID.
func (s *PomodoroService) Delete(_ context.Context, id string) error {
	err := s.storage.DeletePomodoro(id)
	if err != nil {
		return fmt.Errorf("failed to delete pomodoro: %w", err)
	}

	return nil
}

// ActivePomodoro retrieves the current active pomodoro session if any.
func (s *PomodoroService) ActivePomodoro() (*Pomodoro, error) {
	pomodoro, err := s.storage.GetActivePomodoro()
	if err != nil {
		return nil, fmt.Errorf("failed to get active pomodoro: %w", err)
	}

	if pomodoro == nil {
		return nil, fmt.Errorf("no active pomodoro found")
	}

	return s.storagePomodoroToCore(pomodoro), nil
}

// LatestPomodoro retrieves the most recent pomodoro session.
func (s *PomodoroService) LatestPomodoro() (*Pomodoro, error) {
	pomodoro, err := s.storage.GetLatestPomodoro()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest pomodoro: %w", err)
	}

	if pomodoro == nil {
		//nolint:nilnil
		return nil, nil
	}

	return s.storagePomodoroToCore(pomodoro), nil
}

// startTimer starts the timer for a pomodoro session.
func (s *PomodoroService) startTimer(ctx context.Context, id string, duration time.Duration) {
	s.stopTimer()

	s.ticker = time.NewTicker(1 * time.Second)

	originalDuration := duration

	go func() {
		remainingSecs := int(duration.Seconds())

		for {
			select {
			case <-s.ticker.C:
				remainingSecs--

				pomodoro, err := s.storage.UpdatePomodoroState(
					id,
					storage.PomodoroStateActive,
					remainingSecs,
					int(originalDuration.Seconds())-remainingSecs,
				)
				if err != nil {
					log.FromContext(ctx).Error(err, "Failed to update pomodoro time")
				}

				s.publishPomodoroEvent(event.PomodoroTick, pomodoro)

				if remainingSecs <= 0 {
					pomodoro, err := s.storage.UpdatePomodoroState(
						id,
						storage.PomodoroStateFinished,
						0,
						int(originalDuration.Seconds()),
					)
					if err != nil {
						log.FromContext(ctx).Error(err, "Failed to update pomodoro state")
					}

					s.publishPomodoroEvent(event.PomodoroCompleted, pomodoro)

					s.stopTimer()
					return
				}

			case <-s.stopChan:
				return
			}
		}
	}()
}

// stopTimer stops any running timer.
func (s *PomodoroService) stopTimer() {
	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = nil

		select {
		case s.stopChan <- struct{}{}:
		default:
		}
	}
}

// publishPomodoroEvent publishes a pomodoro event to the event bus.
func (s *PomodoroService) publishPomodoroEvent(eventType event.EventType, p *storage.Pomodoro) {
	e := event.PomodoroEvent{
		BaseEvent: event.BaseEvent{
			Type:      eventType,
			Timestamp: time.Now(),
		},
		ID:            p.ID,
		State:         event.PomodoroState(p.State),
		RemainingTime: p.RemainingTime,
		ElapsedTime:   p.ElapsedTime,
		TaskID:        p.TaskID,
		Phase:         event.PomodoroPhase(p.Phase),
		PhaseCount:    p.PhaseCount,
	}

	s.eventBus.Publish(e)
}

// storagePomodoroToCore converts a storage.Pomodoro to a core.Pomodoro.
func (s *PomodoroService) storagePomodoroToCore(p *storage.Pomodoro) *Pomodoro {
	if p == nil {
		return nil
	}

	return &Pomodoro{
		ID:            p.ID,
		State:         event.PomodoroState(p.State),
		StartTime:     p.StartTime,
		WorkDuration:  p.WorkDuration,
		BreakDuration: p.BreakDuration,
		RemainingTime: p.RemainingTime,
		ElapsedTime:   p.ElapsedTime,
		Phase:         event.PomodoroPhase(p.Phase),
		PhaseCount:    p.PhaseCount,
		TaskID:        p.TaskID,
	}
}

func (s *PomodoroService) determinePhaseAndDuration(
	latestPomodoro *Pomodoro,
	workDuration, breakDuration,
	longBreakDuration time.Duration,
) (storage.PomodoroPhase, time.Duration, int) {
	if latestPomodoro == nil {
		return storage.PomodoroPhaseWork, workDuration, 1
	}

	phaseCount := latestPomodoro.PhaseCount + 1

	breakPhases := []event.PomodoroPhase{event.PomodoroPhaseShortBreak, event.PomodoroPhaseLongBreak}
	if slices.Contains(breakPhases, latestPomodoro.Phase) {
		return storage.PomodoroPhaseWork, workDuration, phaseCount
	}

	if time.Duration(phaseCount)%6 == 0 {
		return storage.PomodoroPhaseLongBreak, longBreakDuration, phaseCount
	}

	return storage.PomodoroPhaseShortBreak, breakDuration, phaseCount
}
