// Package task manage task
package task

import (
	"fmt"

	"github.com/gdamore/tcell"

	"github.com/hatappi/gomodoro/task/screen"
)

// Task represents task
type Task struct {
	Name string
}

// Tasks is array of Task
type Tasks []*Task

// GetTaskNames gets task names in array
func (ts Tasks) GetTaskNames() []string {
	var tn []string
	for _, t := range ts {
		tn = append(tn, t.Name)
	}

	return tn
}

// GetTask get tasks name
func GetTask(s tcell.Screen) string {
	c := screen.NewClient(s)

	var tasks Tasks
	for i := 0; i < 50; i++ {
		t := &Task{
			Name: fmt.Sprintf("Task %d", i),
		}
		tasks = append(tasks, t)
	}

	return c.SelectTask(tasks.GetTaskNames())
}
