package plotbot

import (
	"strings"

	"github.com/tkawachi/hipchat"

	"fmt"
)

type BotReply struct {
	To      string
	Message string
}

type Message struct {
	*hipchat.Message
	BotMentioned bool
	FromUser     *User
	FromRoom     *Room
}

func (msg *Message) IsPrivate() bool {
	return msg.FromRoom == nil
}

func (msg *Message) ContainsAnyCased(strs []string) bool {
	for _, s := range strs {
		if strings.Contains(msg.Body, s) {
			return true
		}
	}
	return false
}

func (msg *Message) ContainsAny(strs []string) bool {
	lowerStr := strings.ToLower(msg.Body)

	for _, s := range strs {
		lowerInput := strings.ToLower(s)

		if strings.Contains(lowerStr, lowerInput) {
			return true
		}
	}
	return false
}

func (msg *Message) Contains(s string) bool {
	lowerStr := strings.ToLower(msg.Body)
	lowerInput := strings.ToLower(s)

	if strings.Contains(lowerStr, lowerInput) {
		return true
	}
	return false
}

func (msg *Message) Reply(s string) *BotReply {
	return &BotReply{
		To:      msg.From,
		Message: s,
	}
}

func (msg *Message) ReplyPrivate(s string) *BotReply {
	return &BotReply{
		To:      msg.FromUser.JID,
		Message: s,
	}
}

func (msg *Message) String() string {
	fromUser := "<unknown>"
	if msg.FromUser != nil {
		fromUser = msg.FromUser.Name
	}
	fromRoom := "<none>"
	if msg.FromRoom != nil {
		fromRoom = msg.FromRoom.Name
	}
	return fmt.Sprintf(`Message{"%s", from_user=%s, from_room=%s, mentioned=%v, private=%v}`, msg.Body, fromUser, fromRoom, msg.BotMentioned, msg.IsPrivate())
}
