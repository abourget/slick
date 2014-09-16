package toxin

import "time"

type Message struct {
	From      *User
	Timestamp time.Time
	Text      string
}
