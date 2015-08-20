package slick

import (
	"testing"
	"time"

	"github.com/nlopes/slack"
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
	u := &slack.User{ID: "a_user"}
	m := &Message{Msg: &slack.Msg{Text: "hello mama"}, FromUser: u}

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
			WithUser: &slack.User{ID: "another_user"},
		}, false},
	}

	for i, el := range tests {
		if defaultFilterFunc(el.c, m) != el.r {
			t.Error("defaultFilterFunc Failed, index ", i)
		}
	}
}
