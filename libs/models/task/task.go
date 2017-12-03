package task

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Task struct {
	Name  string
	IsSet bool
}

func (t *Task) SetName(name string) {
	t.Name = name
	t.IsSet = true
}


type TaskList struct {
	NameList []string
	FilePath string
}

func GetNameList(appDir string) (*TaskList, error) {
	filePath := fmt.Sprintf("%s/tasks.toml", appDir)
	if _, err := os.Stat(filePath); err != nil {
		return nil, nil
	}
	var taskList TaskList
	if _, err := toml.DecodeFile(filePath, &taskList); err != nil {
		return nil, err
	}
	taskList.FilePath = filePath
	return &taskList, nil
}

func (tl *TaskList) Save(taskName string) error {
	var config = map[string]interface{}{
		"NameList": append(tl.NameList, taskName),
	}
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(config); err != nil {
		return err
	}
	// if direcotry does'nt exist, create directory
	if err := os.MkdirAll(filepath.Dir(tl.FilePath), 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(tl.FilePath, buf.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}
