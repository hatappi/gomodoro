// Package conv provides functions for converting between GraphQL and core types.
package conv

import (
	gqlgen "github.com/hatappi/gomodoro/internal/client/graphql/generated"
	"github.com/hatappi/gomodoro/internal/core"
)

// ToCoreTask converts a GraphQL Task to a core Task.
func ToCoreTask(task gqlgen.Task) *core.Task {
	return &core.Task{
		ID:        task.Id,
		Title:     task.Title,
		CreatedAt: task.CreatedAt,
		Completed: task.Completed,
	}
}
