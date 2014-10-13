package plotbot

import (
	"fmt"
	"log"
	"time"
)

type Conversation struct {
	ListenUntil     time.Time
	ListenDuration  time.Duration
	WithUser        *User
	InRoom          *Room
	PrivateOnly     bool
	PublicOnly      bool
	Contains        string
	ContainsAny     []string
	MentionsOnly    bool
	MatchMyMessages bool
	FilterFunc      func(*Conversation, *Message) bool
	HandlerFunc     func(*Conversation, *Message)
	Bot             *Bot

	resetCh chan bool
	doneCh  chan bool
}

func (conv *Conversation) Reply(msg *Message, reply string) {
	log.Println("Replying:", reply)
	conv.Bot.replySink <- msg.Reply(reply)
}
func (conv *Conversation) ReplyMention(msg *Message, reply string) {
	if msg.IsPrivate() {
		conv.Reply(msg, reply)
	} else {
		prefix := ""
		if msg.FromUser != nil {
			prefix = fmt.Sprintf("@%s ", msg.FromUser.MentionName)
		}
		conv.Reply(msg, fmt.Sprintf("%s%s", prefix, reply))
	}

}
func (conv *Conversation) ReplyPrivately(msg *Message, reply string) {
	log.Println("Replying privately:", reply)
	conv.Bot.replySink <- msg.ReplyPrivate(reply)
}
func (conv *Conversation) Close() {
}
func (conv *Conversation) ResetDuration() {
}
func (conv *Conversation) isManaged() bool {
	timeout := conv.timeoutDuration()
	return int64(timeout) != 0
}
func (conv *Conversation) launchManager() {
	timeout := conv.timeoutDuration()
	if int64(timeout) == 0 {
		return
	}
	fmt.Println("BOO", timeout)
}
func (conv *Conversation) timeoutDuration() (timeout time.Duration) {
	if !conv.ListenUntil.IsZero() {
		now := time.Now()
		timeout = conv.ListenUntil.Sub(now)
		if int64(timeout) < 0 {
			timeout = 1 * time.Millisecond
		}
	} else if int64(conv.ListenDuration) != 0 {
		timeout = conv.ListenDuration
	}
	return
}

func (conv *Conversation) checkParams() error {
	if !conv.ListenUntil.IsZero() && int64(conv.ListenDuration) != 0 {
		return fmt.Errorf("Specify `ListenUntil` *or* `ListenDuration`, not both.")
	}

	if conv.PrivateOnly && conv.PublicOnly {
		return fmt.Errorf("`PrivateOnly` and `PublicOnly` are mutually exclusive.")
	}

	if conv.Contains != "" && len(conv.ContainsAny) > 0 {
		return fmt.Errorf("`Contains` and `ContainsAny` are mutually exclusive.")
	}

	if conv.HandlerFunc == nil {
		return fmt.Errorf("Required `HandlerFunc` missing")
	}
	// check exclusivity with `FilterFunc` too ?

	return nil
}

func (conv *Conversation) setupChannels() {
	conv.resetCh = make(chan bool, 10)
	conv.doneCh = make(chan bool, 10)
}

func defaultFilterFunc(conv *Conversation, msg *Message) bool {
	if conv.MentionsOnly && !msg.MentionedMe {
		return false
	}

	if conv.PrivateOnly && !msg.IsPrivate() {
		return false
	}

	if conv.PublicOnly && msg.IsPrivate() {
		return false
	}

	if conv.Contains != "" && !msg.Contains(conv.Contains) {
		return false
	}

	if len(conv.ContainsAny) > 0 && !msg.ContainsAny(conv.ContainsAny) {
		return false
	}

	if conv.WithUser != nil && msg.FromUser != conv.WithUser {
		return false
	}

	if conv.InRoom != nil && msg.FromRoom != conv.InRoom {
		return false
	}

	if !conv.MatchMyMessages && msg.FromMe {
		return false
	}

	return true
}
