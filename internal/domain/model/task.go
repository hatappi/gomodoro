package model

// Task represents a task entity
type Task struct {
	ID        string
	Name      string
	Completed bool
}

// Tasks is a collection of Task items
type Tasks []*Task
