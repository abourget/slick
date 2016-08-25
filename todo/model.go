package todo

import "time"

type Task struct {
	ID          string
	CreatedAt   time.Time
	User        string
	Text        []string
	Closed      bool
	ClosingNote string
	ClosedAt    time.Time
}

type Todo []*Task
