package slick

import (
	"log"
	"time"

	"github.com/nlopes/slack"
)

// Reply
//
type Reply struct {
	*slack.OutgoingMessage
	bot *Bot
}

func (r *Reply) OnReaction(f func(addOrRemove string, emoji string)) {
	// TOOD: listen for reactions for a certain time
}

func (r *Reply) AddReaction(emoji string) *Reply {
	r.bot.addListener(&Listener{
		ListenDuration: 20 * time.Second,
		EventHandlerFunc: func(subListen *Listener, event interface{}) {
			if ev, ok := event.(*slack.AckMessage); ok {
				if ev.ReplyTo == r.ID {
					go r.bot.Slack.AddReaction(emoji, slack.NewRefToMessage(r.Channel, ev.Timestamp))
					subListen.Close()
				}
			}
		},
		TimeoutFunc: func(subListen *Listener) {
			log.Println("Reply.AddReaction Listener dropped, because not corresponding AckMessage was received before timeout")
			subListen.Close()
		},
	})
	return r
}

func (r *Reply) Listen(listen *Listener) error {
	listen.Bot = r.bot

	err := listen.checkParams()
	if err != nil {
		log.Println("Reply.Listen(): Invalid Listener: ", err)
		return err
	}

	return r.bot.Listen(&Listener{
		ListenDuration: 20 * time.Second,
		EventHandlerFunc: func(subListen *Listener, event interface{}) {
			if ev, ok := event.(*slack.AckMessage); ok {
				if ev.ReplyTo == r.ID {
					listen.replyAck = ev
					r.bot.addListener(listen)
					subListen.Close()
				}
			}
		},
		TimeoutFunc: func(subListen *Listener) {
			log.Println("Reply.Listen Listener dropped, because not corresponding AckMessage was received before timeout")
			subListen.Close()
		},
	})
}
