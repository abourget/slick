package slick

import (
	"fmt"
	"sync"
)

// UpdateableReply is a Reply that the bot sent, and that it is able
// to update after the fact.
type UpdateableReply struct {
	reply *Reply
	lock  sync.Mutex

	msgTimestamp string // TS from Slack, uniquely identifying our reply.

	newMessage string     // newMessage holds the new message we want to send, and will be dispatched upon reception of the Ack message if set.
	updateMode updateMode // where/how to add/replace the message
}

func (u *UpdateableReply) dispatch() {
	u.lock.Lock()
	defer u.lock.Unlock()

	if u.msgTimestamp == "" {
		return
	}

	if u.newMessage != "" {
		u.reply.bot.Slack.UpdateMessage(u.reply.OutgoingMessage.Channel, u.msgTimestamp, u.newFormattedMessage())
		u.newMessage = ""
	}
}

func (u *UpdateableReply) UpdateSuffix(format string, v ...interface{}) {
	u.updateWithMode(updateSuffix, format, v...)
}

func (u *UpdateableReply) UpdatePrefix(format string, v ...interface{}) {
	u.updateWithMode(updatePrefix, format, v...)
}

func (u *UpdateableReply) Update(format string, v ...interface{}) {
	u.updateWithMode(updateWhole, format, v...)
}

func (u*UpdateableReply) newFormattedMessage() string {
	prevMessage := u.reply.OutgoingMessage.Text
	switch u.updateMode {
	case updateSuffix:
		return fmt.Sprintf("%s%s", prevMessage, u.newMessage)
	case updatePrefix:
		return fmt.Sprintf("%s%s", u.newMessage, prevMessage)
	case updateWhole:
		return u.newMessage
	default:
		panic("there's no other modes !")
	}
}

func (u *UpdateableReply) updateWithMode(mode updateMode, format string, v ...interface{}) {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.newMessage = fmt.Sprintf(format, v...)
	u.updateMode = mode

	go u.dispatch()
}


type updateMode int

const (
	updateSuffix updateMode = iota
	updatePrefix
	updateWhole
)
