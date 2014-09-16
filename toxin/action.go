package toxin

type Action struct {
	ID       string
	AddedBy  *User
	Text     string
	Plusplus []*Plusplus
}

func (action *Action) RecordPlusplus(user *User) {
	pp := NewPlusplus(user)
	action.Plusplus = append(action.Plusplus, pp)
}
