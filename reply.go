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
	r.OnAck(func(ev *slack.AckMessage) {
		go r.bot.Slack.AddReaction(emoji, slack.NewRefToMessage(r.Channel, ev.Timestamp))
	})
	return r
}

func (r *Reply) OnAck(f func(ack *slack.AckMessage)) {
	r.bot.Listen(&Listener{
		ListenDuration: 20 * time.Second,
		EventHandlerFunc: func(subListen *Listener, event interface{}) {
			if ev, ok := event.(*slack.AckMessage); ok {
				if ev.ReplyTo == r.ID {
					f(ev)
					subListen.Close()
				}
			}
		},
		TimeoutFunc: func(subListen *Listener) {
			log.Println("OnAck Listener dropped, because no corresponding AckMessage was received before timeout")
			subListen.Close()
		},
	})
}

// Listen here on Reply is the same as Bot.Listen except that
// ReplyAck() will be filled with the slack.AckMessage before any
// event is dispatched to this listener.
func (r *Reply) Listen(listen *Listener) error {
	listen.Bot = r.bot

	err := listen.checkParams()
	if err != nil {
		log.Println("Reply.Listen(): Invalid Listener: ", err)
		return err
	}

	r.OnAck(func(ev *slack.AckMessage) {
		listen.replyAck = ev
		r.bot.addListener(listen)
	})

	return nil
}

func (r *Reply) DeleteAfter(duration string) *Reply {
	timeDur := parseAutodestructDuration("DeleteAfter", duration)

	r.OnAck(func(ev *slack.AckMessage) {
		go func() {
			time.Sleep(timeDur)
			r.bot.Slack.DeleteMessage(r.Channel, ev.Timestamp)
		}()
	})

	return r
}

func parseAutodestructDuration(funcName string, duration string) time.Duration {
	timeDur, err := time.ParseDuration(duration)
	if err != nil {
		log.Printf("error: %s called with invalid `duration`: %q, using 1 second instead.\n", funcName, duration)
		timeDur = 1 * time.Second
	}
	return timeDur
}
