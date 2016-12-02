package todo

import (
	"strings"
	"time"
)

type Todo []*Task

type byID Todo

func (a byID) Len() int           { return len(a) }
func (a byID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byID) Less(i, j int) bool { return a[i].ID < a[j].ID }

type Task struct {
	ID          string
	CreatedAt   time.Time
	CreatedBy   string
	Text        []string
	Closed      bool
	ClosingNote string
}

func (t *Task) String() string {
	out := "`" + t.ID + "` "
	text := strings.Join(t.Text, " // ")

	if t.Closed {
		out += "~" + text + "~"
	} else {
		out += text
	}

	if t.ClosingNote != "" {
		out += " _" + t.ClosingNote + "_"
	}

	return out
}
