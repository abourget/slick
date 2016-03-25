package slick

import (
	"time"

	"github.com/nlopes/slack"
)

type ReactionListener struct {
	ListenUntil    time.Time
	ListenDuration time.Duration
	FromUser       *slack.User
	Emoji          string
	Type           reaction

	HandlerFunc func(listen *ReactionListener, event *ReactionEvent)
	TimeoutFunc func(*ReactionListener)

	listener *Listener
}

func (reactListen *ReactionListener) newListener() *Listener {
	newListen := &Listener{}
	if !reactListen.ListenUntil.IsZero() {
		newListen.ListenUntil = reactListen.ListenUntil
	}
	if reactListen.ListenDuration != time.Duration(0) {
		newListen.ListenDuration = reactListen.ListenDuration
	}
	if reactListen.TimeoutFunc != nil {
		newListen.TimeoutFunc = func(listen *Listener) {
			reactListen.TimeoutFunc(reactListen)
		}
	}
	reactListen.listener = newListen

	return newListen
}

func (listen *ReactionListener) filterReaction(re *ReactionEvent) bool {
	if listen.Emoji != "" && re.Emoji != listen.Emoji {
		return false
	}
	if listen.FromUser != nil && re.User != listen.FromUser.ID {
		return false
	}
	if int(listen.Type) != 0 && re.Type != listen.Type {
		return false
	}
	return true
}

func (listen *ReactionListener) Close() {
	listen.listener.Close()
}

func (listen *ReactionListener) ResetNewDuration(d time.Duration) {
	listen.listener.ListenDuration = d
	listen.listener.ResetDuration()
}

func (listen *ReactionListener) ResetDuration() {
	listen.listener.ResetDuration()
}

type ReactionEvent struct {
	// Type can be `ReactionAdded` or `ReactionRemoved`
	Type      reaction
	User      string
	Emoji     string
	Timestamp time.Time
	Item      slack.ReactedItem

	// Original objects regarding the reaction, when called on a `Reply`.
	OriginalReply      *Reply
	OriginalAckMessage *slack.AckMessage

	// When called on `Message`
	OriginalMessage *Message

	// Listener is a reference to the thing listening for incoming Reactions
	// you can call .Close() on it after a certain amount of time or after
	// the user you were interested in processed its things.
	Listener *ReactionListener
}

type reaction int

// ReactionAdded is used as the `Type` field of `ReactionEvent` (which
// you can register with `Reply.OnReaction()`)
const ReactionAdded = reaction(2)

// ReactionRemoved is the flipside of `ReactionAdded`.
const ReactionRemoved = reaction(1)


func ParseReactionEvent(event interface{}) *ReactionEvent {
	var re ReactionEvent
	switch ev := event.(type) {
	case *slack.ReactionAddedEvent:
		re.Type = ReactionAdded
		re.Emoji = ev.Reaction
		re.User = ev.User
		re.Item = ev.Item
		re.Timestamp = ev.EventTimestamp.Time()

	case *slack.ReactionRemovedEvent:
		re.Type = ReactionRemoved
		re.Emoji = ev.Reaction
		re.User = ev.User
		re.Item = ev.Item
		re.Timestamp = ev.EventTimestamp.Time()

	default:
		return nil
	}

	return &re
}
