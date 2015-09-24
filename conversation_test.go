package slick

import (
	"regexp"
	"testing"
	"time"

	"github.com/nlopes/slack"
)

func TestListenerCheckParams(t *testing.T) {
	c := Listener{
		ListenUntil:    time.Now(),
		ListenDuration: 120 * time.Second,
	}
	err := c.checkParams()
	if err == nil {
		t.Error("checkParams shouldn't be nil")
	}
}

func TestDefaultFilter(t *testing.T) {
	c := &Listener{}
	u := &slack.User{ID: "a_user"}
	m := &Message{Msg: &slack.Msg{Text: "hello mama"}, FromUser: u}

	if c.filterMessage(m) != true {
		t.Error("filterMessage Failed")
	}

	type El struct {
		c *Listener
		r bool
	}
	tests := []El{
		El{&Listener{}, true},

		El{&Listener{
			Contains: "moo",
		}, false},

		El{&Listener{
			Contains: "MAMA",
		}, true},

		El{&Listener{
			FromUser: u,
		}, true},

		El{&Listener{
			Matches: regexp.MustCompile(`hello`),
		}, true},

		El{&Listener{
			Matches: regexp.MustCompile(`other-message`),
		}, false},

		El{&Listener{
			FromUser: &slack.User{ID: "another_user"},
		}, false},
	}

	for i, el := range tests {
		if el.c.filterMessage(m) != el.r {
			t.Error("filterMessage Failed, index ", i)
		}
	}
}

func TestMatchesMessage(t *testing.T) {
	c := &Listener{Matches: regexp.MustCompile(`(this) (is) (good)`)}
	m := &Message{Msg: &slack.Msg{Text: "yeah this is good and all"}}

	if c.filterMessage(m) != true {
		t.Error("filterMessage Failed")
	}

	if len(m.Match) != 4 {
		t.Error("didn't find 4 matches")
	}

	if m.Match[0] != "this is good" {
		t.Error("didn't find 'this is good'")
	}

	if m.Match[1] != "this" {
		t.Error("didn't find 'this'")
	}
}
