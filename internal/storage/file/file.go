// Package file provides a file-based implementation of the storage interfaces
package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mitchellh/go-homedir"

	"github.com/hatappi/gomodoro/internal/storage"
)

// FileStorage implements storage.Storage using local JSON files.
type FileStorage struct {
	pomodoroFile string
	tasksFile    string
	lockFile     string
	lockHandle   *os.File // File handle for lock file
	mu           sync.Mutex
}

// NewFileStorage creates a new file storage instance.
func NewFileStorage(baseDir string) (storage.Storage, error) {
	if baseDir == "" {
		home, err := homedir.Dir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}

		baseDir = filepath.Join(home, ".gomodoro")
	}

	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	pomodoroFile := filepath.Join(baseDir, "pomodoro.json")
	tasksFile := filepath.Join(baseDir, "tasks.json")
	lockFile := filepath.Join(baseDir, "gomodoro.lock")

	return &FileStorage{
		pomodoroFile: pomodoroFile,
		tasksFile:    tasksFile,
		lockFile:     lockFile,
	}, nil
}

func (f *FileStorage) lock() error {
	if f.lockHandle != nil {
		return nil
	}

	lockFile, err := os.OpenFile(f.lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)

	if os.IsExist(err) {
		return fmt.Errorf("failed to acquire lock, file is locked by another process")
	}

	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}

	if _, err = lockFile.WriteString(fmt.Sprintf("%d", os.Getpid())); err != nil {
		lockFile.Close()
		_ = os.Remove(f.lockFile)
		return fmt.Errorf("failed to write to lock file: %w", err)
	}

	f.lockHandle = lockFile
	return nil
}

func (f *FileStorage) unlock() error {
	if f.lockHandle == nil {
		return nil
	}

	err := f.lockHandle.Close()

	if rmErr := os.Remove(f.lockFile); rmErr != nil {
		if err == nil {
			err = rmErr
		}
	}

	f.lockHandle = nil

	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	return nil
}

func (f *FileStorage) withFileLock(fn func() error) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.lock(); err != nil {
		return err
	}
	defer f.unlock()

	return fn()
}

// SavePomodoro persists a pomodoro to the file.
func (f *FileStorage) SavePomodoro(pomodoro *storage.Pomodoro) error {
	return f.withFileLock(func() error {
		data, err := json.MarshalIndent(pomodoro, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal pomodoro: %w", err)
		}

		if err := os.WriteFile(f.pomodoroFile, data, 0o644); err != nil {
			return fmt.Errorf("failed to write pomodoro file: %w", err)
		}

		return nil
	})
}

// GetLatestPomodoro retrieves the latest pomodoro.
func (f *FileStorage) GetLatestPomodoro() (*storage.Pomodoro, error) {
	var pomodoro *storage.Pomodoro

	err := f.withFileLock(func() error {
		if _, err := os.Stat(f.pomodoroFile); os.IsNotExist(err) {
			return nil
		}

		data, err := os.ReadFile(f.pomodoroFile)
		if err != nil {
			return fmt.Errorf("failed to read pomodoro file: %w", err)
		}

		if len(data) == 0 {
			return nil
		}

		var p storage.Pomodoro
		if err := json.Unmarshal(data, &p); err != nil {
			return fmt.Errorf("failed to unmarshal pomodoro: %w", err)
		}

		pomodoro = &p

		return nil
	})

	return pomodoro, err
}

// GetActivePomodoro retrieves the current active pomodoro.
func (f *FileStorage) GetActivePomodoro() (*storage.Pomodoro, error) {
	var pomodoro *storage.Pomodoro

	err := f.withFileLock(func() error {
		if _, err := os.Stat(f.pomodoroFile); os.IsNotExist(err) {
			return nil
		}

		data, err := os.ReadFile(f.pomodoroFile)
		if err != nil {
			return fmt.Errorf("failed to read pomodoro file: %w", err)
		}

		if len(data) == 0 {
			return nil
		}

		var p storage.Pomodoro
		if err := json.Unmarshal(data, &p); err != nil {
			return fmt.Errorf("failed to unmarshal pomodoro: %w", err)
		}

		if p.State == storage.PomodoroStateActive || p.State == storage.PomodoroStatePaused {
			pomodoro = &p
		}

		return nil
	})

	return pomodoro, err
}

// UpdatePomodoroState updates the state and remaining time of a pomodoro.
func (f *FileStorage) UpdatePomodoroState(id string, state storage.PomodoroState, remainSec int, elapsedSec int) (*storage.Pomodoro, error) {
	var pomodoro *storage.Pomodoro

	err := f.withFileLock(func() error {
		if _, err := os.Stat(f.pomodoroFile); os.IsNotExist(err) {
			return fmt.Errorf("no active pomodoro found")
		}

		data, err := os.ReadFile(f.pomodoroFile)
		if err != nil {
			return fmt.Errorf("failed to read pomodoro file: %w", err)
		}

		if err := json.Unmarshal(data, &pomodoro); err != nil {
			return fmt.Errorf("failed to unmarshal pomodoro: %w", err)
		}

		if pomodoro.ID != id {
			return fmt.Errorf("pomodoro ID mismatch")
		}

		pomodoro.State = state
		pomodoro.RemainingTime = time.Duration(remainSec) * time.Second
		pomodoro.ElapsedTime = time.Duration(elapsedSec) * time.Second

		updatedData, err := json.MarshalIndent(pomodoro, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal updated pomodoro: %w", err)
		}

		if err := os.WriteFile(f.pomodoroFile, updatedData, 0o644); err != nil {
			return fmt.Errorf("failed to write updated pomodoro file: %w", err)
		}

		return nil
	})

	return pomodoro, err
}

// DeletePomodoro deletes the pomodoro session file if the ID matches the current session.
func (f *FileStorage) DeletePomodoro(id string) error {
	return f.withFileLock(func() error {
		if _, err := os.Stat(f.pomodoroFile); os.IsNotExist(err) {
			return nil
		}

		data, err := os.ReadFile(f.pomodoroFile)
		if err != nil {
			return fmt.Errorf("failed to read pomodoro file: %w", err)
		}

		if len(data) == 0 {
			return nil
		}

		var p storage.Pomodoro
		if err := json.Unmarshal(data, &p); err != nil {
			return fmt.Errorf("failed to unmarshal pomodoro: %w", err)
		}

		if p.ID != id {
			return nil
		}

		if err := os.Remove(f.pomodoroFile); err != nil {
			return fmt.Errorf("failed to delete pomodoro file: %w", err)
		}
		return nil
	})
}

// SaveTask persists a task to the tasks file.
func (f *FileStorage) SaveTask(task *storage.Task) error {
	return f.withFileLock(func() error {
		tasks, err := f.readTasks()
		if err != nil {
			return err
		}

		found := false
		for i, t := range tasks {
			if t.ID == task.ID {
				tasks[i] = task
				found = true
				break
			}
		}

		if !found {
			tasks = append(tasks, task)
		}

		return f.writeTasks(tasks)
	})
}

// GetTasks retrieves all tasks.
func (f *FileStorage) GetTasks() ([]*storage.Task, error) {
	var tasks []*storage.Task

	err := f.withFileLock(func() error {
		var err error
		tasks, err = f.readTasks()
		return err
	})

	return tasks, err
}

// GetTaskByID retrieves a specific task by ID.
func (f *FileStorage) GetTaskByID(id string) (*storage.Task, error) {
	var foundTask *storage.Task

	err := f.withFileLock(func() error {
		tasks, err := f.readTasks()
		if err != nil {
			return err
		}

		for _, task := range tasks {
			if task.ID == id {
				foundTask = task
				return nil
			}
		}

		return fmt.Errorf("task with ID %s not found", id)
	})

	return foundTask, err
}

// UpdateTask updates an existing task.
func (f *FileStorage) UpdateTask(task *storage.Task) error {
	return f.withFileLock(func() error {
		tasks, err := f.readTasks()
		if err != nil {
			return err
		}

		found := false
		for i, t := range tasks {
			if t.ID == task.ID {
				tasks[i] = task
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("task with ID %s not found", task.ID)
		}

		return f.writeTasks(tasks)
	})
}

func (f *FileStorage) DeleteTask(id string) error {
	return f.withFileLock(func() error {
		tasks, err := f.readTasks()
		if err != nil {
			return err
		}

		foundIndex := -1
		for i, task := range tasks {
			if task.ID == id {
				foundIndex = i
				break
			}
		}

		if foundIndex == -1 {
			return fmt.Errorf("task with ID %s not found", id)
		}

		tasks = append(tasks[:foundIndex], tasks[foundIndex+1:]...)

		return f.writeTasks(tasks)
	})
}

func (f *FileStorage) readTasks() ([]*storage.Task, error) {
	if _, err := os.Stat(f.tasksFile); os.IsNotExist(err) {
		return make([]*storage.Task, 0), nil
	}

	data, err := os.ReadFile(f.tasksFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks file: %w", err)
	}

	if len(data) == 0 {
		return make([]*storage.Task, 0), nil
	}

	var tasks []*storage.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks: %w", err)
	}

	return tasks, nil
}

func (f *FileStorage) writeTasks(tasks []*storage.Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	if err := os.WriteFile(f.tasksFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write tasks file: %w", err)
	}

	return nil
}
