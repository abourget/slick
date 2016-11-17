package todo

import (
	"strings"
	"time"
)

type Todo []*Task

type Task struct {
	ID          string
	CreatedAt   time.Time
	CreatedBy   string
	Text        []string
	Closed      bool
}

func (t *Task) String() string {
	out := "`" + t.ID + "` "
	text := strings.Join(t.Text, " // ")

	out += text

	return out
}
