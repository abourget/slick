package wicked

import "time"

type Decision struct {
	ID        string
	Timestamp time.Time
	AddedBy   *User
	Text      string
	Plusplus  []*Plusplus
}

func (decision *Decision) RecordPlusplus(user *User) {
	pp := NewPlusplus(user)
	decision.Plusplus = append(decision.Plusplus, pp)
}
