package task

import (
	"bufio"
	"fmt"
	"os"

	taskModel "github.com/hatappi/gomodoro/src/models/task"
	"github.com/hatappi/gomodoro/src/selector"
	"github.com/hatappi/gomodoro/src/task"
)

func Get(tasks []string) (*taskModel.Task, error) {
	var (
		selectTask *taskModel.Task
		err        error
	)

	if len(tasks) > 0 {
		selectTask, err = selector.Task(tasks)
		if err != nil || !selectTask.IsSet {
			return selectTask, err
		}
	}

	for {
		if selectTask.Name == "" {
			fmt.Print("Please Input New Task > ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			selectTask.Name = scanner.Text()
			err := task.Save(append(tasks, selectTask.Name))
			if err != nil {
				return selectTask, err
			}
		} else {
			break
		}
	}
	return selectTask, err
}
