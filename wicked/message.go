package wicked

import "time"

type Message struct {
	From      *User
	Timestamp time.Time
	Text      string
}
