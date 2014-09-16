package toxin

import (
	"fmt"
	"strings"
	"time"

	"github.com/plotly/plotbot"
)

func init() {
	plotbot.RegisterStringList("toxin annoyments", []string{
		"friends, told you it was enough",
		"can't you guys control yourselves!",
		"I'm going to have to intervene!",
		"underestimated or undisciplied ?",
		"I can't believe I need to repeat this",
		"really!",
		"you've gone way overboard",
		"time is precious my friends, time is precious",
		"/me can't believe what he's seeing!",
		"someone! step up and tell the others they need to wrap up!",
		"are you listening ?",
	})
}

type Subject struct {
	ID        string
	AddedBy   *User
	Text      string
	TimeLimit time.Duration

	// Global timing
	Duration  time.Duration // Sum of (EndTime - StartTime) as they happen
	BeginTime time.Time
	FinalTime time.Time
	// Local timing, when we talk about the subject
	StartTime time.Time
	EndTime   time.Time

	Actions     []*Action
	Plusplus    []*Plusplus
	Refs        []*Reference
	WasExtended bool

	// Conversation management
	doneCh chan bool
}

func (subject *Subject) Timebox() string {
	if subject.TimeLimit == 0 {
		return "[no limit]"
	} else {
		return strings.Replace(subject.TimeLimit.String(), "0s", "", 1)
	}
}

func (subject *Subject) RecordPlusplus(user *User) {
	pp := NewPlusplus(user)
	subject.Plusplus = append(subject.Plusplus, pp)
}

func (subject *Subject) String() string {
	return fmt.Sprintf("%s (timebox: %s ref: s#%s)", subject.Text, subject.Timebox(), subject.ID)
}

func (subject *Subject) Start(bot *plotbot.Bot, msg *plotbot.Message) {
	if subject.BeginTime.IsZero() {
		subject.BeginTime = time.Now()
	}
	subject.StartTime = time.Now()

	if subject.doneCh == nil {
		subject.doneCh = make(chan bool, 5)
		go subject.manageSubject(bot, msg)
	}
}

func (subject *Subject) Stop() {
	subject.EndTime = time.Now()
	subject.Duration += subject.EndTime.Sub(subject.StartTime)
	subject.FinalTime = time.Now()

	if subject.doneCh != nil {
		close(subject.doneCh)
	}
	subject.doneCh = nil
}

func (subject *Subject) manageSubject(bot *plotbot.Bot, msg *plotbot.Message) {
	doneCh := subject.doneCh
	durationAt90Percent := time.Duration(float64(subject.TimeLimit) * 0.90)
	annoymentMinutes := (int(subject.TimeLimit.Minutes()/30) + 1) * 2
	annoymentInterval := time.Duration(time.Duration(annoymentMinutes) * time.Minute)

	select {
	case <-time.After(durationAt90Percent):
		bot.Reply(msg, fmt.Sprintf("Ok folks, %s left before subject s#%s times out", subject.beforeTimeoutString(), subject.ID))
	case <-doneCh:
		return
	}

	select {
	case <-time.After(subject.remainingBeforeTimeout()):
		bot.Reply(msg, fmt.Sprintf(`Time's up! Hit "!next" to start discussing the next subject.`))
	case <-doneCh:
		return
	}

	for {
		select {
		case <-time.After(annoymentInterval):
			bot.Reply(msg, plotbot.RandomString("toxin annoyments"))
		case <-doneCh:
			return
		}
	}
}

func (subject *Subject) beforeTimeoutString() string {
	remainingSecs := subject.remainingBeforeTimeout().Seconds()
	if remainingSecs < 60.0 {
		return "less than a minute"
	} else {
		mins := int(remainingSecs+1.0) % 60
		if mins == 1 {
			return fmt.Sprintf("one minute")
		} else {
			return fmt.Sprintf("%d minutes")
		}
	}
}

func (subject *Subject) remainingBeforeTimeout() time.Duration {
	t := time.Now()
	elapsed := t.Sub(subject.StartTime)
	remaining := subject.TimeLimit - elapsed
	return remaining
}
