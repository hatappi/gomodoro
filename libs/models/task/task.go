package task

type Task struct {
	Name  string
	IsSet bool
}

func (t *Task) SetName(name string) {
	t.Name = name
	t.IsSet = true
}
