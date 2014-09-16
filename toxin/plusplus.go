package toxin

type Plusplus struct {
	From *User
}

func NewPlusplus(from *User) *Plusplus {
	pp := &Plusplus{
		From: from,
	}
	return pp
}
