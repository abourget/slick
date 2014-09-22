package wicked

import "time"

type Reference struct {
	AddedBy   *User
	Timestamp time.Time
	URL       string
	Text      string
}
