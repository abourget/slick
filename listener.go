package slick

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/nlopes/slack"
)

type Listener struct {
	// replyAck is filled when you call Listen() on a Reply.
	replyAck *slack.AckMessage

	// ListenUntil sets an absolute date at which this Listener
	// expires and stops listening.  ListenUntil and ListenDuration
	// are optional and mutually exclusive.
	ListenUntil time.Time

	// ListenDuration sets a timeout Duration, after which this
	// Listener stops listening and is garbage collected.  A call
	// to `ResetTimeout()` restarts the listening period for another
	// `ListenDuration`.
	ListenDuration time.Duration

	// FromUser filters out incoming messages that are not with
	// `*User` (publicly or privately)
	FromUser *slack.User
	// FromChannel filters messages that are sent to a different room than
	// `Room`. This can be mixed and matched with `FromUser`
	FromChannel *Channel

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

	// Matches checks that the given text matches the given Regexp
	// with a `FindStringSubmatch` call. It will set the `Message.Match`
	// attribute.
	Matches *regexp.Regexp

	// ListenForEdits will trigger a message when a user edits a
	// message as well as creates a new one.
	ListenForEdits bool

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
	//
	// Also, if you override TimeoutFunc, you need to call Close() yourself
	// otherwise, the conversation is not removed from the listeners

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

// ReplyAck returns the AckMessage received that corresponds to the Reply
// on which you called Listen()
func (listen *Listener) ReplyAck() *slack.AckMessage {
	return listen.replyAck
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
		return fmt.Errorf("specify `ListenUntil` *or* `ListenDuration`, not both.")
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

func (listen *Listener) filterAndDispatchMessage(msg *Message) {
	if listen.filterMessage(msg) {
		listen.MessageHandlerFunc(listen, msg)
	}
}

// filterMessage applies checks from a Listener against a Message.
func (listen *Listener) filterMessage(msg *Message) bool {
	if msg.Msg.SubType == "message_deleted" {
		// like "message_deleted"
		return false
	}
	if msg.Msg.SubType == "message_changed" && !listen.ListenForEdits {
		return false
	}

	// Never pick up on other bot's messages
	if msg.Msg.SubType == "bot_message" {
		return false
	}

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

	if listen.Matches != nil {
		match := listen.Matches.FindStringSubmatch(msg.Text)
		msg.Match = match
		if match == nil {
			return false
		}
	}

	// If there is no msg.FromUser, the message is filtered out.
	if listen.FromUser != nil && (msg.FromUser == nil || msg.FromUser.ID != listen.FromUser.ID) {
		return false
	}

	if listen.FromChannel != nil {
		if msg.FromChannel == nil {
			return false
		}
		if msg.FromChannel.ID != listen.FromChannel.ID {
			return false
		}
	}

	if !listen.MatchMyMessages && msg.FromMe {
		return false
	}

	return true
}
