package slick

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

type botReply struct {
	To   string
	Text string
}

type Message struct {
	*slack.Msg
	SubMessage    *slack.Msg
	bot           *Bot
	MentionsMe    bool
	IsEdition     bool
	FromMe        bool
	FromUser      *slack.User
	FromChannel   *slack.Channel
}

func (msg *Message) IsPrivate() bool {
	return msg.Channel == ""
}

func (msg *Message) ContainsAnyCased(strs []string) bool {
	for _, s := range strs {
		if strings.Contains(msg.Text, s) {
			return true
		}
	}
	return false
}

func (msg *Message) ContainsAny(strs []string) bool {
	lowerStr := strings.ToLower(msg.Text)

	for _, s := range strs {
		lowerInput := strings.ToLower(s)

		if strings.Contains(lowerStr, lowerInput) {
			return true
		}
	}
	return false
}

func (msg *Message) ContainsAll(strs []string) bool {

	lowerStr := strings.ToLower(msg.Text)

	for _, s := range strs {
		lowerInput := strings.ToLower(s)

		if !strings.Contains(lowerStr, lowerInput) {
			return false
		}
	}
	return true
}

func (msg *Message) Contains(s string) bool {
	lowerStr := strings.ToLower(msg.Text)
	lowerInput := strings.ToLower(s)

	if strings.Contains(lowerStr, lowerInput) {
		return true
	}
	return false
}

func (msg *Message) HasPrefix(prefix string) bool {
	return strings.HasPrefix(msg.Text, prefix)
}

func (msg *Message) AddReaction(emoticon string) {
	msg.bot.Slack.AddReaction(emoticon, slack.NewRefToMessage(msg.Channel, msg.Timestamp))
}

func (msg *Message) RemoveReaction(emoticon string) {
	msg.bot.Slack.RemoveReaction(emoticon, slack.NewRefToMessage(msg.Channel, msg.Timestamp))
}

func (msg *Message) Reply(s string) *slack.OutgoingMessage {
	to := msg.User
	if msg.Channel != "" {
		to = msg.Channel
	}
	return msg.bot.SendOutgoingMessage(s, to)
}

func (msg *Message) ReplyPrivately(s string) *slack.OutgoingMessage {
	return msg.bot.SendOutgoingMessage(s, msg.User)
}

// ReplyMention replies with a @mention named prefixed, when replying
// in public. When replying in private, nothing is added.
func (msg *Message) ReplyMention(reply string) *slack.OutgoingMessage {
	if msg.IsPrivate() {
		return msg.Reply(reply)
	}
	prefix := ""
	if msg.FromUser != nil {
		prefix = fmt.Sprintf("<@%s> ", msg.FromUser.Name)
	}
	return msg.Reply(fmt.Sprintf("%s%s", prefix, reply))
}

// ReplyFlash sends a reply like "Reply", but will self-destruct after `duration`
// time (as a time.ParseDuration).
func (msg *Message) ReplyFlash(duration string, reply string) *slack.OutgoingMessage {
	timeDur := parseAutodestructDuration("ReplyFlash", duration)
	outMsg := msg.Reply(reply)
	msg.selfDestruct(timeDur, outMsg)
	return outMsg
}

// ReplyMentionFlash sends a reply like "Reply", but will self-destruct after `duration`
// time (as a time.ParseDuration).
func (msg *Message) ReplyMentionFlash(duration string, reply string) *slack.OutgoingMessage {
	timeDur := parseAutodestructDuration("ReplyMentionFlash", duration)
	outMsg := msg.ReplyMention(reply)
	msg.selfDestruct(timeDur, outMsg)
	return outMsg
}

func (msg *Message) selfDestruct(duration time.Duration, outMsg *slack.OutgoingMessage) {
	msg.bot.ListenFor(&Listener{
		ListenDuration: time.Duration(30 * time.Second), // before the ACK
		EventHandlerFunc: func(listen *Listener, event interface{}) {
			if ev, ok := event.(*slack.AckMessage); ok {
				if ev.ReplyTo == outMsg.ID {
					go func() {
						<-time.After(duration)
						msg.bot.Slack.DeleteMessage(outMsg.Channel, ev.Timestamp)
					}()
				}
			}
		},
	})
}

func parseAutodestructDuration(funcName string, duration string) time.Duration {
	timeDur, err := time.ParseDuration(duration)
	if err != nil {
		log.Printf("%s called with invalid `duration`: %q, using 1 second instead.\n", funcName, duration)
		timeDur = time.Duration(1 * time.Second)
	}
	return timeDur
}

func (msg *Message) String() string {
	return fmt.Sprintf("%#v", msg)
}

func (msg *Message) applyMentionsMe(bot *Bot) {
	if msg.IsPrivate() {
		msg.MentionsMe = true
	}

	m := reAtMention.FindStringSubmatch(msg.Text)
	if m != nil && m[1] == bot.Myself.ID {
		msg.MentionsMe = true
	}
}

func (msg *Message) applyFromMe(bot *Bot) {
	if msg.User == bot.Myself.ID {
		msg.FromMe = true
	}
}

var reAtMention = regexp.MustCompile(`<@([A-Z0-9]+)(|([^>]+))>`)
