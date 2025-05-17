package conv

import (
	"github.com/hatappi/gomodoro/internal/core"
	"github.com/hatappi/gomodoro/internal/graph/model"
)

// FromCoreTask converts a core.Task to a model.Task.
func FromCoreTask(task *core.Task) *model.Task {
	if task == nil {
		return nil
	}

	return &model.Task{
		ID:        task.ID,
		Title:     task.Title,
		CreatedAt: task.CreatedAt,
		Completed: task.Completed,
	}
}
