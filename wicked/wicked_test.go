package wicked

import (
	"testing"

	"github.com/abourget/slick"
	"github.com/nlopes/slack"
)

func TestFindNextRoom(t *testing.T) {
	c2 := slack.Channel{}
	c2.ID = "room2"
	c2.Name = "room2"
	c3 := slack.Channel{}
	c3.ID = "room3"
	c3.Name = "room3"

	w := &Wicked{
		bot: &slick.Bot{Channels: map[string]slack.Channel{
			"room2": c2,
			"room3": c3,
		}},
		meetings: map[string]*Meeting{
			"room1": &Meeting{},
		},
		confRooms: []string{
			"room1",
			"room2",
			"room3",
		},
	}

	res := w.FindAvailableRoom("other")

	if res == nil {
		t.Fail()
	}
	if res.ID != "room2" {
		t.Error(`Should be "room2"`)
	}

	res = w.FindAvailableRoom("room3")

	if res == nil {
		t.Fail()
	}
	if res.ID != "room3" {
		t.Error(`Should be "room3"`)
	}
}

func TestFindNextRoomNilFromRoom(t *testing.T) {
	c1 := slack.Channel{}
	c1.ID = "room1"
	c1.Name = "room1"
	w := &Wicked{
		bot: &slick.Bot{Channels: map[string]slack.Channel{
			"room1": c1,
		}},
		meetings:  map[string]*Meeting{},
		confRooms: []string{"room1"},
	}

	res := w.FindAvailableRoom("")

	if res == nil {
		t.Fail()
	}
	if res.ID != "room1" {
		t.Error(`Should be "room1"`)
	}
}

func TestFindNextRoomAllTake(t *testing.T) {
	w := &Wicked{
		bot: &slick.Bot{Channels: map[string]slack.Channel{}},
		meetings: map[string]*Meeting{
			"room1": &Meeting{},
		},
		confRooms: []string{"room1"},
	}

	res := w.FindAvailableRoom("other")

	if res != nil {
		t.Error(`Should be nil`)
	}
}
