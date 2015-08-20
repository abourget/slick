package slick

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
)

type BotReply struct {
	To   string
	Text string
}

type Message struct {
	*slack.Msg
	SubMessage  *slack.Msg
	MentionsMe  bool
	IsEdition   bool
	FromMe      bool
	FromUser    *slack.User
	FromChannel *slack.Channel
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

func (msg *Message) Reply(s string) *BotReply {
	rep := &BotReply{
		Text: s,
	}
	if msg.Channel != "" {
		rep.To = msg.Channel
	} else {
		rep.To = msg.User
	}
	return rep
}

func (msg *Message) ReplyPrivately(s string) *BotReply {
	return &BotReply{
		To:   msg.User,
		Text: s,
	}
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
