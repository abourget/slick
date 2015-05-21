package wicked

import (
	"fmt"
	"strings"
	"time"

	"github.com/abourget/slick"
	"github.com/nlopes/slack"
)

func init() {
	slick.RegisterStringList("wicked annoyments", []string{
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

type Meeting struct {
	ID           string
	CreatedBy    *User
	Channel      string
	ChannelID    string
	Goal         string
	TimeLimit    time.Duration
	StartTime    time.Time
	EndTime      time.Time
	Logs         []*Message
	Decisions    []*Decision
	Refs         []*Reference
	Participants []*User

	sendToRoom func(string)
	setTopic   func(string)
	doneCh     chan bool
}

func NewMeeting(id string, user *slack.User, goal string, bot *slick.Bot, channel *slack.Channel, uuidNow time.Time) *Meeting {
	meeting := &Meeting{}
	meeting.ID = id
	meeting.Channel = channel.Name
	meeting.ChannelID = channel.Id
	meeting.Goal = strings.TrimSpace(goal)
	meeting.StartTime = uuidNow
	meeting.Decisions = []*Decision{}
	meeting.Refs = []*Reference{}
	meeting.Logs = []*Message{}
	meeting.Participants = []*User{}
	meeting.sendToRoom = func(msg string) {
		bot.SendToChannel(meeting.ChannelID, msg)
	}
	meeting.setTopic = func(topic string) {
		// TODO: set a topic with Slack.
		//hipchatv2.SetTopic(bot.Config.HipchatApiToken, roomId, topic)
	}

	newUser := meeting.ImportUser(user)
	meeting.CreatedBy = newUser

	return meeting
}

func (meeting *Meeting) ImportUser(user *slack.User) *User {
	fromEmail := user.Profile.Email

	for _, user := range meeting.Participants {
		if user.Email == fromEmail {
			return user
		}
	}

	newUser := &User{
		Email:    user.Profile.Email,
		Fullname: user.Name,
		PhotoURL: user.Profile.Image48,
	}

	meeting.Participants = append(meeting.Participants, newUser)

	return newUser
}

func (meeting *Meeting) AddDecision(user *User, text string, uuidNow time.Time) *Decision {
	id := meeting.NextDecisionID()

	decision := &Decision{
		ID:        id,
		Timestamp: uuidNow,
		AddedBy:   user,
		Text:      text,
	}

	meeting.Decisions = append(meeting.Decisions, decision)

	return decision
}

func (meeting *Meeting) GetDecisionByID(id string) *Decision {
	for _, decision := range meeting.Decisions {
		if decision.ID == id {
			return decision
		}
	}
	return nil
}

func (meeting *Meeting) AddReference(user *User, text string, uuidNow time.Time) *Reference {
	ref := &Reference{
		AddedBy:   user,
		Timestamp: uuidNow,
	}
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "http") {
		chunks := strings.SplitN(text, " ", 2)
		ref.URL = chunks[0]
		if len(chunks) > 1 {
			ref.Text = chunks[1]
		}
	} else {
		ref.Text = text
	}

	meeting.Refs = append(meeting.Refs, ref)

	return ref
}

func (meeting *Meeting) NextDecisionID() string {
	for i := 1; i < 1000; i++ {
		strID := fmt.Sprintf("%d", i)
		taken := false
		for _, decision := range meeting.Decisions {
			if decision.ID == strID {
				taken = true
				break
			}
		}
		if !taken {
			return strID
		}
	}
	return "fail"
}

func (meeting *Meeting) Conclude() {
	meeting.EndTime = time.Now()
	// TODO: liberate the current "Wicked Confroom"
}
