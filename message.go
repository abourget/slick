package slick

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
)

type botReply struct {
	To   string
	Text string
}

type Message struct {
	*slack.Msg
	SubMessage  *slack.Msg
	bot         *Bot
	MentionsMe  bool
	IsEdit      bool
	FromMe      bool
	FromUser    *slack.User
	FromChannel *Channel

	// Match contains the result of
	// Listener.Matches.FindStringSubmatch(msg.Text), when `Matches`
	// is set on the `Listener`.
	Match []string
}

func (msg *Message) IsPrivate() bool {
	return strings.HasPrefix(msg.Channel, "D")
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

func (msg *Message) AddReaction(emoticon string) *Message {
	msg.bot.Slack.AddReaction(emoticon, slack.NewRefToMessage(msg.Channel, msg.Timestamp))
	return msg
}

func (msg *Message) RemoveReaction(emoticon string) *Message {
	msg.bot.Slack.RemoveReaction(emoticon, slack.NewRefToMessage(msg.Channel, msg.Timestamp))
	return msg
}

func (msg *Message) ListenReaction(reactListen *ReactionListener) {
	msg.bot.ListenReaction(msg.Timestamp, reactListen)
}

func (msg *Message) Reply(text string, v ...interface{}) *Reply {
	to := msg.User
	if msg.Channel != "" {
		to = msg.Channel
	}
	text = Format(text, v...)
	return msg.bot.SendOutgoingMessage(text, to)
}

func (msg *Message) ReplyPrivately(text string, v ...interface{}) *Reply {
	text = Format(text, v...)
	return msg.bot.SendPrivateMessage(msg.User, text)
}

// ReplyMention replies with a @mention named prefixed, when replying
// in public. When replying in private, nothing is added.
func (msg *Message) ReplyMention(text string, v ...interface{}) *Reply {
	if msg.IsPrivate() {
		return msg.Reply(text, v...)
	}
	prefix := ""
	if msg.FromUser != nil {
		prefix = fmt.Sprintf("<@%s> ", msg.FromUser.Name)
	}
	return msg.Reply(fmt.Sprintf("%s%s", prefix, text), v...)
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
	if msg.User != "" && msg.User == bot.Myself.ID {
		msg.FromMe = true
	}
}

var reAtMention = regexp.MustCompile(`<@([A-Z0-9]+)(|([^>]+))>`)

// Format conditionally formats using fmt.Sprintf if there is more
// than one argument, otherwise returns the first parameter
// uninterpreted.
func Format(s string, v ...interface{}) string {
	count := len(v)
	if count == 0 {
		return s
	}
	return fmt.Sprintf(s, v...)
}
