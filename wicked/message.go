package wicked

import "time"

// TODO: add highlight
type Message struct {
	From      *User
	Timestamp time.Time
	Text      string
}
