package wicked

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/plotly/plotbot"
)

type Wicked struct {
	bot          *plotbot.Bot
	meetings     map[string]*Meeting
	pastMeetings []*Meeting
}

func init() {
	plotbot.RegisterPlugin(&Wicked{})
}

func (wicked *Wicked) InitChatPlugin(bot *plotbot.Bot) {
	wicked.bot = bot
	wicked.meetings = make(map[string]*Meeting)
}

var config = &plotbot.ChatPluginConfig{
	EchoMessages: false,
	OnlyMentions: false,
}

func (wicked *Wicked) ChatConfig() *plotbot.ChatPluginConfig {
	return config
}

func (wicked *Wicked) ChatHandler(bot *plotbot.Bot, msg *plotbot.Message) {
	room := msg.FromRoom.JID
	meeting, meetingExists := wicked.meetings[room]
	if strings.HasPrefix(msg.Body, "!wicked ") {
		if meetingExists {
			bot.Reply(msg, "Wicked meeting already running in current room")
			return
		}
		// Accept !wicked  and start.

		id := wicked.NextMeetingID()
		meeting := NewMeeting(id, msg.FromUser, msg.Body[6:], room)
		wicked.pastMeetings = append(wicked.pastMeetings, meeting)
		wicked.meetings[room] = meeting
		bot.Reply(msg, fmt.Sprintf("Wicked meeting started.  Welcome aboard.  Access report at %s/wicked/%s", wicked.bot.Config.WebBaseURL, meeting.ID))
	}

	if !meetingExists {
		if strings.HasPrefix(msg.Body, "!subjects") {
			bot.ReplyPrivate(msg, fmt.Sprintf(`No meeting running in "%s"`, msg.FromRoom.Name))
		}

		return
	}

	//
	// Manage messages from within a meeting
	//
	user := meeting.ImportUser(msg.FromUser)

	if strings.HasPrefix(msg.Body, "!subject ") {
		subject := meeting.AddSubject(user, msg.Body[9:])
		if subject == nil {
			bot.Reply(msg, "Whoops, wrong syntax for !subject")
		} else {
			bot.Reply(msg, fmt.Sprintf("Subject added, timebox: %s, ref: s#%s", subject.Timebox(), subject.ID))
		}

	} else if strings.HasPrefix(msg.Body, "!subjects") {
		go func() {
			bot.ReplyPrivate(msg, fmt.Sprintf(`Subjects in "%s":`, msg.FromRoom.Name))
			for i, subject := range meeting.Subjects {
				time.Sleep(100 * time.Millisecond)
				current := ""
				if subject == meeting.CurrentSubject {
					current = " <---- currently discussing"
				}
				bot.ReplyPrivate(msg, fmt.Sprintf("%d. %s %s", i+1, subject, current))
			}
		}()

	} else if strings.HasPrefix(msg.Body, "!start") {

		if meeting.CurrentSubject != nil {
			subject := meeting.CurrentSubject
			bot.Reply(msg, fmt.Sprintf("Hmm, you've started already. We're discussing s#%s: %s", subject.ID, subject.Text))
		} else {
			if len(meeting.Subjects) == 0 {
				bot.Reply(msg, fmt.Sprintf("No subjects listed, add some with !subject"))
			} else {
				subject := meeting.NextSubject(bot, msg)
				bot.Reply(msg, fmt.Sprintf("Goal: %s\nStarting subject: %s", meeting.Goal, subject))
			}
		}

	} else if strings.HasPrefix(msg.Body, "!next") {
		if len(meeting.Subjects) == 0 {
			bot.Reply(msg, "No subjects listed, add some with !subject")
		} else {

			// Wrap up counters
			// Start the next subject, the first is none is started.
			if meeting.CurrentSubject == nil {
				meeting.NextSubject(bot, msg)
				subject := meeting.CurrentSubject
				bot.Reply(msg, fmt.Sprintf("Goal: %s\nStarting subject: %s", meeting.Goal, subject))
			} else {
				if meeting.CurrentIsLast() {
					bot.Reply(msg, fmt.Sprintf("No more subjects.  Add some with !subject or !conclude the wicked meeting"))
				} else {
					subject := meeting.NextSubject(bot, msg)
					bot.Reply(msg, fmt.Sprintf("Goal: %s\nPassing on to subject: %s", meeting.Goal, subject))
				}
			}
		}

	} else if strings.HasPrefix(msg.Body, "!previous") {
		if !wicked.ensureOnSubject(meeting, msg) {
			return
		}

	} else if strings.HasPrefix(msg.Body, "!extend ") {
		if !wicked.ensureOnSubject(meeting, msg) {
			return
		}

		duration, err := time.ParseDuration(msg.Body[8:])
		if err != nil {
			bot.Reply(msg, "Hmm, wrong syntax !extend, or invalid duration")
		} else {
			subject := meeting.CurrentSubject
			subject.WasExtended = true
			subject.TimeLimit = duration
			bot.Reply(msg, fmt.Sprintf("Extended for another %s, don't do that too often!", subject.Timebox()))
		}

		// Extend counters, update timers and notifications

	} else if strings.HasPrefix(msg.Body, "!action ") {

		if !wicked.ensureOnSubject(meeting, msg) {
			return
		}
		// Add to Subject *and* Meeting
		action := meeting.AddAction(user, meeting.CurrentSubject, msg.Body[8:])
		if action == nil {
			bot.Reply(msg, "Whoops, wrong syntax for !action")
		} else {
			bot.Reply(msg, fmt.Sprintf("Action added, ref: a#%s", action.ID))
		}

	} else if strings.HasPrefix(msg.Body, "!ref ") {

		meeting.AddReference(user, msg.Body[5:])
		bot.Reply(msg, "Ref. added")

	} else if strings.HasPrefix(msg.Body, "!conclude") {
		meeting.Conclude()
		// TODO: kill all waiting goroutines dealing with messaging
		delete(wicked.meetings, room)
		bot.Reply(msg, "Concluding wicked meeting, that's all folks!")

	} else if match := actionMatcher.FindStringSubmatch(msg.Body); match != nil {

		action := meeting.GetActionByID(match[1])
		if action != nil {
			if match[2] == "++" {
				action.RecordPlusplus(user)
				bot.ReplyMention(msg, "noted")
			}
		}

	} else if match := subjectMatcher.FindStringSubmatch(msg.Body); match != nil {

		subject := meeting.GetSubjectByID(match[1])
		if subject != nil {
			if match[2] == "++" {
				subject.RecordPlusplus(user)
				bot.ReplyMention(msg, "noted")
			}
		}
	}

	// Log message
	newMessage := &Message{
		From:      user,
		Timestamp: time.Now(),
		Text:      msg.Body,
	}
	meeting.Logs = append(meeting.Logs, newMessage)
	/**
	* Handle everything for this meeting:
	*
	* !wicked [goal]
	* !subject [s#tag] <duration as \d+m> <Subject text>
	* !next
	* !previous
	* !extend
	* !action [a#tag] <Action text>
	* !ref [url] <Some reference text>
	* !conclude
	*
	* Handles: a#tag++, s#tag++ in any sentence
	**/
}

func (wicked *Wicked) NextMeetingID() string {
	for i := 1; i < 10000; i++ {
		strID := fmt.Sprintf("%d", i)
		taken := false
		for _, meeting := range wicked.pastMeetings {
			if meeting.ID == strID {
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

var actionMatcher = regexp.MustCompile(`a#([a-z]+|\d+)(\+\+)?`)
var subjectMatcher = regexp.MustCompile(`s#([a-z]+|\d+)(\+\+)?`)

func (wicked *Wicked) ensureOnSubject(meeting *Meeting, msg *plotbot.Message) bool {
	if meeting.CurrentSubject == nil {
		wicked.bot.Reply(msg, "We haven't started a subject yet, start with !next")
		return false
	}
	return true
}
