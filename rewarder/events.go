package rewarder

import (
	"time"

	"github.com/plotly/plotbot"
)

func (rew *Rewarder) LogEvent(user *plotbot.User, event string, data interface{}) error {
	return nil
}

func (rew *Rewarder) FetchLogsSince(user *plotbot.User, since time.Time, event string, data interface{}) error {
	return nil
}

func (rew *Rewarder) FetchLastLog(user *plotbot.User, event string, data interface{}) error {
	return nil
}

func (rew *Rewarder) FetchLastNLogs(user *plotbot.User, num int, event string, data interface{}) error {
	return nil
}
