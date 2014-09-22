package wicked

import (
	"testing"

	"github.com/plotly/plotbot"
)

func TestFindNextRoom(t *testing.T) {
	w := &Wicked{
		bot: &plotbot.Bot{Rooms: []plotbot.Room{
			plotbot.Room{JID: "room2"},
			plotbot.Room{JID: "room3"},
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
	if res.JID != "room2" {
		t.Error(`Should be "room2"`)
	}

	res = w.FindAvailableRoom("room3")

	if res == nil {
		t.Fail()
	}
	if res.JID != "room3" {
		t.Error(`Should be "room3"`)
	}
}

func TestFindNextRoomNilFromRoom(t *testing.T) {
	w := &Wicked{
		bot: &plotbot.Bot{Rooms: []plotbot.Room{
			plotbot.Room{JID: "room1"},
		}},
		meetings:  map[string]*Meeting{},
		confRooms: []string{"room1"},
	}

	res := w.FindAvailableRoom("")

	if res == nil {
		t.Fail()
	}
	if res.JID != "room1" {
		t.Error(`Should be "room1"`)
	}
}

func TestFindNextRoomAllTake(t *testing.T) {
	w := &Wicked{
		bot: &plotbot.Bot{Rooms: []plotbot.Room{}},
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
