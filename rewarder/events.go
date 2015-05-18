package rewarder

import (
	"encoding/json"
	"time"

	"github.com/abourget/slick"
)

type Event struct {
	UserEmail string
	Name      string
	Timestamp time.Time
	Data      string
}

func (rew *Rewarder) LogEvent(user *slick.User, event string, data interface{}) error {
	userLogs, ok := rew.logs[user.Email]
	if !ok {
		userLogs = make([]*Event, 0)
	}

	ev := &Event{
		UserEmail: user.Email,
		Name:      event,
		Timestamp: time.Now().UTC(),
		Data:      rew.serializeData(data),
	}

	// TODO: Flush to Redis
	userLogs = append(userLogs, ev)

	rew.logs[user.Email] = userLogs

	return nil
}

func (rew *Rewarder) serializeData(v interface{}) string {
	res, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(res)
}

func (rew *Rewarder) unserializeData(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

func (rew *Rewarder) FetchEventsSince(user *slick.User, since time.Time, event string, data interface{}) error {
	return nil
}

func (rew *Rewarder) FetchLastEvent(user *slick.User, event string, data interface{}) error {
	return nil
}

func (rew *Rewarder) FetchLastNEvents(user *slick.User, num int, event string, data interface{}) error {
	return nil
}
