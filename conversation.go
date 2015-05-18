package slick

import (
	"fmt"
	"log"
	"time"
)

type Conversation struct {
	// ListenUntil sets an absolute date at which this Conversation
	// expires and stops listening.  ListenUntil and ListenDuration
	// are optional and mutually exclusive.
	ListenUntil time.Time
	// ListenDuration sets a timeout Duration, after which this
	// Conversation stops listening and is garbage collected.  A call
	// to `ResetTimeout()` restarts the listening period for another
	// `ListenDuration`.
	ListenDuration time.Duration

	// WithUser filters out incoming messages that are not with
	// `*User` (publicly or privately)
	WithUser *User
	// InRoom filters messages that are sent to a different room than
	// `Room`. This can be mixed and matched with `WithUser`
	InRoom *Room

	// PrivateOnly filters out public messages.
	PrivateOnly bool
	// PublicOnly filters out private messages.  Mutually exclusive
	// with `PrivateOnly`.
	PublicOnly bool

	// Contains checks whether the `string` is in the message body
	// (after lower-casing both components).
	Contains string

	// ContainsAny checks that any one of the specified strings exist
	// as substrings in the message body.  Mutually exclusive with
	// `Contains`.
	ContainsAny []string

	// MentionsMe filters out messages that do not mention the Bot's
	// `bot.Config.MentionName`
	MentionsMeOnly bool

	// MatchMyMessages equal to false filters out messages that the bot
	// himself sent.
	MatchMyMessages bool

	// FilterFunc is run with each message to verify whether to call
	// `HandlerFunc` with the message.  See `defaultFilterFunc`
	FilterFunc func(*Conversation, *Message) bool

	// HandlerFunc is the main handling function, which receives messages
	// to handle.
	HandlerFunc func(*Conversation, *Message)

	// TimeoutFunc is called when a conversation expires after
	// `ListenDuration` or `ListenUntil` delays.  It is *not* called
	// if you explicitly call `Close()` on the conversation, or if
	// you did not set `ListenDuration` nor `ListenUntil`.
	TimeoutFunc func(*Conversation)

	// Ref to the bot instance.  Populated for you when `ListenFor`-ing the
	// Conversation.
	Bot *Bot

	resetCh chan bool
	doneCh  chan bool
}

func (conv *Conversation) Reply(msg *Message, reply string) {
	conv.Bot.Reply(msg, reply)
}

func (conv *Conversation) ReplyMention(msg *Message, reply string) {
	conv.Bot.ReplyMention(msg, reply)
}

func (conv *Conversation) ReplyPrivately(msg *Message, reply string) {
	conv.Bot.ReplyPrivately(msg, reply)
}

// Close terminates the Conversation management goroutine, and stops
// any further listening and message handling
func (conv *Conversation) Close() {
	conv.Bot.delConversationCh <- conv
	conv.doneCh <- true
}

// ResetDuration re-initializes the timeout set by
// `Conversation.ListenDuration`, and continues listening for another
// such duration.
func (conv *Conversation) ResetDuration() error {
	if int64(conv.ListenDuration) == 0 {
		msg := "Conversation has no ListenDuration"
		log.Println("ResetDuration() error: ", msg)
		return fmt.Errorf(msg)
	}

	conv.resetCh <- true

	return nil
}

func (conv *Conversation) isManaged() bool {
	timeout := conv.timeoutDuration()
	return int64(timeout) != 0
}

func (conv *Conversation) launchManager() {
	for {
		timeout := conv.timeoutDuration()

		select {
		case <-time.After(timeout):
			if conv.TimeoutFunc != nil {
				conv.TimeoutFunc(conv)
			}
			return
		case <-conv.resetCh:
			continue
		case <-conv.doneCh:
			return
		}
	}
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
	if conv.MentionsMeOnly && !msg.MentionsMe {
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

	if conv.WithUser != nil && msg.FromUser.JID != conv.WithUser.JID {
		return false
	}

	if conv.InRoom != nil {
		if msg.FromRoom == nil {
			return false
		}
		if msg.FromRoom.JID != conv.InRoom.JID {
			return false
		}
	}

	if !conv.MatchMyMessages && msg.FromMe {
		return false
	}

	return true
}
