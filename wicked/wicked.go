package wicked

/**
 * TODO:
 * Implement notion of Wicked Confroom, and its management
 * Remove "Subject" altogether
 * Implement !join , with Wicked meetings references W11 and W22, etc..
 * Change Plusplus to D12++ and R23++ and W22++ ..
 * Time reminder, simply send to the Wicked Confroom, a reminer of the time,
 *   maybe as a bold statement, for how long the meeting has ran, and if it's
 *   over the time.
 */

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/abourget/slick"
	"github.com/abourget/slack"
)

type Wicked struct {
	bot          *slick.Bot
	confRooms    []string
	meetings     map[string]*Meeting
	pastMeetings []*Meeting
}

func init() {
	slick.RegisterPlugin(&Wicked{})
}

func (wicked *Wicked) InitChatPlugin(bot *slick.Bot) {
	wicked.bot = bot
	wicked.meetings = make(map[string]*Meeting)

	var conf struct {
		Wicked struct {
			Confrooms []string `json:"conf_rooms"`
		}
	}
	bot.LoadConfig(&conf)
	for _, confroom := range conf.Wicked.Confrooms {
		wicked.confRooms = append(wicked.confRooms, confroom)
	}

	bot.ListenFor(&slick.Conversation{
		HandlerFunc: wicked.ChatHandler,
	})
}

func (wicked *Wicked) ChatHandler(conv *slick.Conversation, msg *slick.Message) {
	bot := conv.Bot
	uuidNow := time.Now()

	if strings.HasPrefix(msg.Text, "!wicked ") {
		fromRoom := ""
		if msg.FromChannel != nil {
			fromRoom = msg.FromChannel.Id
		}

		availableRoom := wicked.FindAvailableRoom(fromRoom)

		if availableRoom == nil {
			conv.Reply(msg, "No available Wicked Confroom for a meeting! Seems you'll need to create new Wicked Confrooms !")
			goto continueLogging
		}

		id := wicked.NextMeetingID()
		meeting := NewMeeting(id, msg.FromUser, msg.Text[7:], bot, availableRoom, uuidNow)

		wicked.pastMeetings = append(wicked.pastMeetings, meeting)
		wicked.meetings[availableRoom.Id] = meeting

		if availableRoom.Id == fromRoom {
			meeting.sendToRoom(fmt.Sprintf(`*** Starting wicked meeting W%s in here.`, meeting.ID))
		} else {
			conv.Reply(msg, fmt.Sprintf(`*** Starting wicked meeting W%s in room "%s". Join with !join W%s`, meeting.ID, availableRoom.Name, meeting.ID))
			initiatedFrom := ""
			if fromRoom != "" {
				initiatedFrom = fmt.Sprintf(` in "%s"`, msg.FromChannel.Name)
			}
			meeting.sendToRoom(fmt.Sprintf(`*** Wicked meeting initiated by @%s%s. Goal: %s`, msg.FromUser.Name, initiatedFrom, meeting.Goal))
		}

		meeting.sendToRoom(fmt.Sprintf(`*** Access report at %s/wicked/%s.html`, wicked.bot.Config.WebBaseURL, meeting.ID))
		meeting.setTopic(fmt.Sprintf(`[Running] W%s goal: %s`, meeting.ID, meeting.Goal))
	} else if strings.HasPrefix(msg.Text, "!join") {
		match := joinMatcher.FindStringSubmatch(msg.Text)
		if match == nil {
			conv.ReplyMention(msg, `invalid !join syntax. Use something like "!join W123"`)
		} else {
			for _, meeting := range wicked.meetings {
				if match[1] == meeting.ID {
					meeting.sendToRoom(fmt.Sprintf(`*** @%s asked to join`, msg.FromUser.Name))
				}
			}
		}
	}

continueLogging:

	//
	// Public commands and messages
	//
	if msg.FromChannel == nil {
		return
	}
	room := msg.FromChannel.Id
	meeting, meetingExists := wicked.meetings[room]
	if !meetingExists {
		return
	}

	user := meeting.ImportUser(msg.FromUser)

	if strings.HasPrefix(msg.Text, "!proposition ") {
		decision := meeting.AddDecision(user, msg.Text[12:], uuidNow)
		if decision == nil {
			conv.Reply(msg, "Whoops, wrong syntax for !proposition")
		} else {
			conv.Reply(msg, fmt.Sprintf("Proposition added, ref: D%s", decision.ID))
		}

	} else if strings.HasPrefix(msg.Text, "!ref ") {

		meeting.AddReference(user, msg.Text[4:], uuidNow)
		conv.Reply(msg, "Ref. added")

	} else if strings.HasPrefix(msg.Text, "!conclude") {
		meeting.Conclude()
		// TODO: kill all waiting goroutines dealing with messaging
		delete(wicked.meetings, room)
		meeting.sendToRoom("Concluding Wicked meeting, that's all folks!")
		meeting.setTopic(fmt.Sprintf(`[Concluded] W%s goal: %s`, meeting.ID, meeting.Goal))

	} else if match := decisionMatcher.FindStringSubmatch(msg.Text); match != nil {

		decision := meeting.GetDecisionByID(match[1])
		if decision != nil {
			decision.RecordPlusplus(user)
			conv.ReplyMention(msg, "noted")
		}

	}

	// Log message
	newMessage := &Message{
		From:      user,
		Timestamp: uuidNow,
		Text:      msg.Text,
	}
	meeting.Logs = append(meeting.Logs, newMessage)
}

func (wicked *Wicked) FindAvailableRoom(fromRoom string) *slack.Channel {
	nextFree := ""
	for _, confRoom := range wicked.confRooms {
		_, occupied := wicked.meetings[confRoom]
		if occupied {
			continue
		}
		if fromRoom == confRoom {
			return wicked.bot.GetChannelByName(confRoom)
		}
		if nextFree == "" {
			nextFree = confRoom
		}
	}

	return wicked.bot.GetChannelByName(nextFree)
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

var decisionMatcher = regexp.MustCompile(`(?mi)D(\d+)\+\+`)
var joinMatcher = regexp.MustCompile(`!join\s+(?mi)W(\d+)`)
