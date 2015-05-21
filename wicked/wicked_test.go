package wicked

import (
	"testing"

	"github.com/abourget/slick"
	"github.com/nlopes/slack"
)

func TestFindNextRoom(t *testing.T) {
	w := &Wicked{
		bot: &slick.Bot{Channels: map[string]slack.Channel{
			"room2": {BaseChannel: slack.BaseChannel{Id: "room2"}, Name: "room2"},
			"room3": {BaseChannel: slack.BaseChannel{Id: "room3"}, Name: "room3"},
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
	if res.Id != "room2" {
		t.Error(`Should be "room2"`)
	}

	res = w.FindAvailableRoom("room3")

	if res == nil {
		t.Fail()
	}
	if res.Id != "room3" {
		t.Error(`Should be "room3"`)
	}
}

func TestFindNextRoomNilFromRoom(t *testing.T) {
	w := &Wicked{
		bot: &slick.Bot{Channels: map[string]slack.Channel{
			"room1": {BaseChannel: slack.BaseChannel{Id: "room1"}, Name: "room1"},
		}},
		meetings:  map[string]*Meeting{},
		confRooms: []string{"room1"},
	}

	res := w.FindAvailableRoom("")

	if res == nil {
		t.Fail()
	}
	if res.Id != "room1" {
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
