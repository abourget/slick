package blaster

import "time"

type try struct {
	start  time.Time
	end    time.Time
	status int
}

func (t *try) Delta() time.Duration {
	if t.start.IsZero() || t.end.IsZero() {
		return 0
	}
	return t.end.Sub(t.start)
}
