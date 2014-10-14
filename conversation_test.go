package plotbot

import (
	"testing"
	"time"

	"github.com/tkawachi/hipchat"
)

func TestConversationCheckParams(t *testing.T) {
	c := Conversation{
		ListenUntil:    time.Now(),
		ListenDuration: 120 * time.Second,
	}
	err := c.checkParams()
	if err == nil {
		t.Error("checkParams shouldn't be nil")
	}
}

func TestDefeaultFilter(t *testing.T) {
	c := &Conversation{}
	u := &User{JID: "a_user"}
	m := &Message{Message: &hipchat.Message{Body: "hello mama"}, FromUser: u}

	if defaultFilterFunc(c, m) != true {
		t.Error("defaultFilterFunc Failed")
	}

	type El struct {
		c *Conversation
		r bool
	}
	tests := []El{
		El{&Conversation{}, true},

		El{&Conversation{
			Contains: "moo",
		}, false},

		El{&Conversation{
			Contains: "MAMA",
		}, true},

		El{&Conversation{
			WithUser: u,
		}, true},

		El{&Conversation{
			WithUser: &User{JID: "another_user"},
		}, false},

	}

	for i, el := range tests {
		if defaultFilterFunc(el.c, m) != el.r {
			t.Error("defaultFilterFunc Failed, index ", i)
		}
	}
}
