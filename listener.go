package slick

import (
	"fmt"
	"log"
	"time"

	"github.com/nlopes/slack"
)

type Listener struct {
	// OutgoingMessage can be set to a *slack.OutgoingMessage. If it is
	// non-nil, slick will listen for the Ack and add the AckMessage
	// to OutgoingMessageAck *before* really attaching this listener.
	//
	OutgoingMessage *slack.OutgoingMessage

	// OutgoingMessageAck is fill by the bot when you specify OutgoingMessage
	// before calling `Listen`
	OutgoingMessageAck *slack.AckMessage

	// ListenUntil sets an absolute date at which this Listener
	// expires and stops listening.  ListenUntil and ListenDuration
	// are optional and mutually exclusive.
	ListenUntil time.Time

	// ListenDuration sets a timeout Duration, after which this
	// Listener stops listening and is garbage collected.  A call
	// to `ResetTimeout()` restarts the listening period for another
	// `ListenDuration`.
	ListenDuration time.Duration

	// WithUser filters out incoming messages that are not with
	// `*User` (publicly or privately)
	WithUser *slack.User
	// InChannel filters messages that are sent to a different room than
	// `Room`. This can be mixed and matched with `WithUser`
	InChannel *slack.Channel

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

	// MessageHandlerFunc is a handling function provided by the user, and
	// called when a relevant message comes in.
	MessageHandlerFunc func(*Listener, *Message)

	// EventHandlerFunc is a handling function provided by the user, and
	// called when any event is received. These messages are dispatched
	// to each Listener in turn, after the bot has processed it.
	// If the event is a Message, then the `slick.Message` will be non-nil.
	//
	// When receiving a `*slack.MessageEvent`, slick will wrap it in a `*slick.Message`
	// which embeds the the original event, but adds quite a few functionalities, like
	// reply modes, etc..
	EventHandlerFunc func(*Listener, interface{})

	// TimeoutFunc is called when a conversation expires after
	// `ListenDuration` or `ListenUntil` delays.  It is *not* called
	// if you explicitly call `Close()` on the conversation, or if
	// you did not set `ListenDuration` nor `ListenUntil`.
	TimeoutFunc func(*Listener)

	// Bot is a reference to the bot instance.  It will always be populated before being
	// passed to handler functions.
	Bot *Bot

	resetCh chan bool
	doneCh  chan bool
}

// Close terminates the Listener management goroutine, and stops
// any further listening and message handling
func (listen *Listener) Close() {
	listen.Bot.delListenerCh <- listen
	listen.doneCh <- true
}

// ResetDuration re-initializes the timeout set by
// `Listener.ListenDuration`, and continues listening for another
// such duration.
func (listen *Listener) ResetDuration() error {
	if int64(listen.ListenDuration) == 0 {
		msg := "Listener has no ListenDuration"
		log.Println("ResetDuration() error: ", msg)
		return fmt.Errorf(msg)
	}

	listen.resetCh <- true

	return nil
}

func (listen *Listener) isManaged() bool {
	timeout := listen.timeoutDuration()
	return int64(timeout) != 0
}

func (listen *Listener) launchManager() {
	for {
		timeout := listen.timeoutDuration()

		select {
		case <-time.After(timeout):
			if listen.TimeoutFunc != nil {
				listen.TimeoutFunc(listen)
			}
			return
		case <-listen.resetCh:
			continue
		case <-listen.doneCh:
			return
		}
	}
}
func (listen *Listener) timeoutDuration() (timeout time.Duration) {
	if !listen.ListenUntil.IsZero() {
		now := time.Now()
		timeout = listen.ListenUntil.Sub(now)
		if int64(timeout) < 0 {
			timeout = 1 * time.Millisecond
		}
	} else if int64(listen.ListenDuration) != 0 {
		timeout = listen.ListenDuration
	}
	return
}

func (listen *Listener) checkParams() error {
	if !listen.ListenUntil.IsZero() && int64(listen.ListenDuration) != 0 {
		return fmt.Errorf("Specify `ListenUntil` *or* `ListenDuration`, not both.")
	}

	if listen.PrivateOnly && listen.PublicOnly {
		return fmt.Errorf("`PrivateOnly` and `PublicOnly` are mutually exclusive.")
	}

	if listen.Contains != "" && len(listen.ContainsAny) > 0 {
		return fmt.Errorf("`Contains` and `ContainsAny` are mutually exclusive.")
	}

	if (listen.MessageHandlerFunc == nil && listen.EventHandlerFunc == nil) || (listen.MessageHandlerFunc != nil && listen.EventHandlerFunc != nil) {
		return fmt.Errorf("One and only one of `MessageHandlerFunc` and `EventHandlerFunc` is required.")
	}

	return nil
}

func (listen *Listener) setupChannels() {
	listen.resetCh = make(chan bool, 10)
	listen.doneCh = make(chan bool, 10)
}

// defaultFilterFunc applies checks from a Listener against a Message.
func defaultFilterFunc(listen *Listener, msg *Message) bool {
	if listen.MentionsMeOnly && !msg.MentionsMe {
		return false
	}

	if listen.PrivateOnly && !msg.IsPrivate() {
		return false
	}

	if listen.PublicOnly && msg.IsPrivate() {
		return false
	}

	if listen.Contains != "" && !msg.Contains(listen.Contains) {
		return false
	}

	if len(listen.ContainsAny) > 0 && !msg.ContainsAny(listen.ContainsAny) {
		return false
	}

	if listen.WithUser != nil && msg.FromUser.ID != listen.WithUser.ID {
		return false
	}

	if listen.InChannel != nil {
		if msg.FromChannel == nil {
			return false
		}
		if msg.FromChannel.ID != listen.InChannel.ID {
			return false
		}
	}

	if !listen.MatchMyMessages && msg.FromMe {
		return false
	}

	return true
}