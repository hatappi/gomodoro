package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.73

import (
	"context"

	"github.com/hatappi/gomodoro/internal/graph/conv"
	"github.com/hatappi/gomodoro/internal/graph/model"
)

// CreateTask is the resolver for the createTask field.
func (r *mutationResolver) CreateTask(ctx context.Context, input model.CreateTaskInput) (*model.Task, error) {
	task, err := r.TaskService.CreateTask(ctx, input.Title)
	if err != nil {
		return nil, err
	}

	return conv.FromCoreTask(task), nil
}

// UpdateTask is the resolver for the updateTask field.
func (r *mutationResolver) UpdateTask(ctx context.Context, input model.UpdateTaskInput) (*model.Task, error) {
	updatedTask, err := r.TaskService.UpdateTask(ctx, input.ID, *input.Title)
	if err != nil {
		return nil, err
	}

	return conv.FromCoreTask(updatedTask), nil
}

// DeleteTask is the resolver for the deleteTask field.
func (r *mutationResolver) DeleteTask(ctx context.Context, id string) (*bool, error) {
	if err := r.TaskService.DeleteTask(ctx, id); err != nil {
		return nil, err
	}

	return conv.ToPointer(true), nil
}

// Tasks is the resolver for the tasks field.
func (r *queryResolver) Tasks(ctx context.Context) (*model.TaskConnection, error) {
	tasks, err := r.TaskService.GetAllTasks()
	if err != nil {
		return nil, err
	}

	var edges []*model.TaskEdge
	for _, task := range tasks {
		tm := conv.FromCoreTask(task)

		edges = append(edges, &model.TaskEdge{
			Cursor: tm.ID,
			Node:   tm,
		})
	}

	pageInfo := &model.PageInfo{
		HasNextPage:     false,
		HasPreviousPage: false,
	}

	if len(tasks) > 0 {
		pageInfo.StartCursor = conv.ToPointer(edges[0].Cursor)
		pageInfo.EndCursor = conv.ToPointer(edges[len(edges)-1].Cursor)
	}

	return &model.TaskConnection{
		Edges:      edges,
		TotalCount: len(tasks),
		PageInfo:   pageInfo,
	}, nil
}

// Task is the resolver for the task field.
func (r *queryResolver) Task(ctx context.Context, id string) (*model.Task, error) {
	task, err := r.TaskService.GetTaskByID(id)
	if err != nil {
		return nil, err
	}
	return conv.FromCoreTask(task), nil
}
