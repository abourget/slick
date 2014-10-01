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

func (decision *Decision) IsProposition() bool {
	if len(decision.Plusplus) > 0 {
		return true
	} else {
		return false
	}
}
