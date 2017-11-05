package task

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/go-homedir"
)

var filePath string

type TaskList struct {
	NameList []string
}

func init() {
	homeDir, err := homedir.Dir()
	if err != nil {
		homeDir = "/opt"
	}
	filePath = homeDir + "/.gomodoro/tasks.yml"
}

func GetNameList() ([]string, error) {
	if _, err := os.Stat(filePath); err != nil {
		return []string{}, nil
	}
	var taskList TaskList
	if _, err := toml.DecodeFile(filePath, &taskList); err != nil {
		return nil, err
	}
	return taskList.NameList, nil
}

func Save(taskList []string) error {
	var config = map[string]interface{}{
		"NameList": taskList,
	}
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(config); err != nil {
		return err
	}
	// if direcotry does'nt exist, create directory
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filePath, buf.Bytes(), os.ModePerm); err != nil {
		return err
	}
	return nil
}
