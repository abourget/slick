package hipchat

import (
	"fmt"
	"github.com/plotly/hipchat/xmpp"
	"log"
	"strings"
	"time"
)

type Delay struct {
	Stamp   time.Time
	FromJid string
}

// A Message represents a message received from HipChat.
type Message struct {
	From        string
	To          string
	Body        string
	MentionName string
	Delay       *Delay
}

// Build Message from xmpp.Element.
// Returns nil when it's not a message.
func NewMessage(elem *xmpp.Element) (msg *Message) {
	if !elem.IsMessage() {
		return nil
	}
	attr := elem.AttrMap()
	msg = new(Message)
	msg.From = attr["from"]
	msg.To = attr["to"]
	body := elem.FindChild(func(el *xmpp.Element) bool {
		return el.Name().Local == "body"
	})
	if body == nil {
		return nil
	}
	msg.Body = body.CharData
	delay := elem.FindChild(func(el *xmpp.Element) bool {
		return el.Name().Local == "delay"
	})
	if delay != nil {
		d := new(Delay)
		delayAttr := delay.AttrMap()
		stampStr := strings.Split(delayAttr["stamp"], " ")[0]
		stamp, err := time.Parse(time.RFC3339, stampStr)
		if err != nil {
			log.Fatal(err)
		}
		d.Stamp = stamp
		d.FromJid = delayAttr["from_jid"]
		msg.Delay = d
	}
	return
}

func (msg *Message) FromNick() string {
	slice := strings.Split(msg.From, "/")
	if len(slice) < 2 {
		return ""
	}
	return slice[1]
}

func (msg *Message) String() string {
	result := fmt.Sprintf(`Message(%v -> %v, "%v"`, msg.From, msg.To, msg.Body)
	if msg.Delay != nil {
		result += fmt.Sprintf(", delay: %v", *msg.Delay)
	}
	result += ")"
	return result
}
