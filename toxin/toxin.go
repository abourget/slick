package toxin

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/abourget/ahipbot"
)

type Toxin struct {
	bot          *ahipbot.Bot
	meetings     map[string]*Meeting
	pastMeetings []*Meeting
	config       *Config
}

type Config struct {
	WebBaseURL string `json:"web_base_url"`
}

func init() {
	ahipbot.RegisterPlugin(func(bot *ahipbot.Bot) ahipbot.Plugin {
		var conf struct {
			Toxin Config
		}
		bot.LoadConfig(&conf)
		return &Toxin{
			bot:      bot,
			meetings: make(map[string]*Meeting),
			config:   &conf.Toxin,
		}
	})
}

var config = &ahipbot.PluginConfig{
	EchoMessages: false,
	OnlyMentions: false,
}

func (toxin *Toxin) Config() *ahipbot.PluginConfig {
	return config
}

func (toxin *Toxin) Handle(bot *ahipbot.Bot, msg *ahipbot.BotMessage) {
	room := msg.FromRoom.JID
	meeting, meetingExists := toxin.meetings[room]
	if strings.HasPrefix(msg.Body, "!toxin ") {
		if meetingExists {
			bot.Reply(msg, "Toxin already running in current room")
			return
		}
		// Accept !toxin  and start.

		id := toxin.NextMeetingID()
		meeting := NewMeeting(id, msg.FromUser, msg.Body[6:], room)
		toxin.pastMeetings = append(toxin.pastMeetings, meeting)
		toxin.meetings[room] = meeting
		bot.Reply(msg, fmt.Sprintf("Toxin started.  Welcome aboard.  Access report at %s/toxin/%s", toxin.config.WebBaseURL, meeting.ID))
	}

	if !meetingExists {
		return
	}

	user := meeting.ImportUser(msg.FromUser)

	if strings.HasPrefix(msg.Body, "!subject ") {
		subject := meeting.AddSubject(user, msg.Body[9:])
		if subject == nil {
			bot.Reply(msg, "Whoops, wrong syntax for !subject")
		} else {
			bot.Reply(msg, fmt.Sprintf("Subject added, timebox: %s, ref: s#%s", subject.Timebox(), subject.ID))
		}

	} else if strings.HasPrefix(msg.Body, "!start") {

		if meeting.CurrentSubject != nil {
			subject := meeting.CurrentSubject
			bot.Reply(msg, fmt.Sprintf("Hmm, you've started already. We're discussing s#%s: %s", subject.ID, subject.Text))
		} else {
			if len(meeting.Subjects) == 0 {
				bot.Reply(msg, fmt.Sprintf("No subjects listed, add some with !subject"))
			} else {
				subject := meeting.NextSubject()
				bot.Reply(msg, fmt.Sprintf("Starting Toxin with subject: %s\n- Time limit: %s - Ref: s#%s", subject.Text, subject.TimeLimit.String(), subject.ID))
			}
		}

	} else if strings.HasPrefix(msg.Body, "!next") {
		if len(meeting.Subjects) == 0 {
			bot.Reply(msg, "No subjects listed, add some with !subject")
		} else {

			// Wrap up counters
			// Start the next subject, the first is none is started.
			if meeting.CurrentSubject == nil {
				meeting.NextSubject()
				subject := meeting.CurrentSubject
				bot.Reply(msg, fmt.Sprintf("Starting Toxin with subject: %s\n- Time limit: %s - Ref: s#%s", subject.Text, subject.TimeLimit.String(), subject.ID))
			} else {
				if meeting.CurrentIsLast() {
					bot.Reply(msg, fmt.Sprintf("No more subjects.  Add some with !subject or !conclude the toxin"))
				} else {
					prevSubject := meeting.CurrentSubject
					subject := meeting.NextSubject()
					bot.Reply(msg, fmt.Sprintf("ok, done with subject s#%s: %s\nDiscussing subject: %s\n- Time limit: %s - Ref: s#%s", prevSubject.ID, prevSubject.Text, subject.Text, subject.TimeLimit.String(), subject.ID))
				}
			}
		}

	} else if strings.HasPrefix(msg.Body, "!previous") {
		if !toxin.ensureOnSubject(meeting, msg) {
			return
		}

	} else if strings.HasPrefix(msg.Body, "!extend ") {
		if !toxin.ensureOnSubject(meeting, msg) {
			return
		}

		duration, err := time.ParseDuration(msg.Body[8:])
		if err != nil {
			bot.Reply(msg, "hmm, wrong syntax !extend, or invalid duration")
		} else {
			subject := meeting.CurrentSubject
			subject.WasExtended = true
			subject.TimeLimit += duration
			bot.Reply(msg, fmt.Sprintf("extended to a total of %s, don't do that too often!", subject.TimeLimit.String()))
		}


		// Extend counters, update timers and notifications

	} else if strings.HasPrefix(msg.Body, "!action ") {

		if !toxin.ensureOnSubject(meeting, msg) {
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
		delete(toxin.meetings, room)
		bot.Reply(msg, "Concluding toxin, that's all folks!")

	} else if match := actionMatcher.FindStringSubmatch(msg.Body); match != nil {

		action := meeting.GetActionByID(match[1])
		if action != nil {
			if match[2] == "++" {
				action.RecordPlusplus(user)
				bot.Reply(msg, "noted")
			}
		}

	} else if match := subjectMatcher.FindStringSubmatch(msg.Body); match != nil {

		subject := meeting.GetSubjectByID(match[1])
		if subject != nil {
			if match[2] == "++" {
				subject.RecordPlusplus(user)
				bot.Reply(msg, "noted")
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
	* !toxin [goal]
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

func (toxin *Toxin) NextMeetingID() string {
	for i := 1; i < 10000; i++ {
		strID := fmt.Sprintf("%d", i)
		taken := false
		for _, meeting := range toxin.pastMeetings {
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

func (toxin *Toxin) ensureOnSubject(meeting *Meeting, msg *ahipbot.BotMessage) bool {
	if meeting.CurrentSubject == nil {
		toxin.bot.Reply(msg, "We haven't started a subject yet, start with !next")
		return false
	}
	return true
}
