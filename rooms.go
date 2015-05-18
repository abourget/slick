package slick

import (
	"strings"

	"github.com/abourget/slick/hipchatv2"
)

type Room struct {
	ID                int64  `json:"id"`
	JID               string `json:"xmpp_jid"`
	Name              string `json:"name"`
	Topic             string `json:"topic"`
	Privacy           string `json:"privacy"`
	IsArchived        bool   `json:"is_archived"`
	IsGuestAccessible bool   `json:"is_guest_accessible"`
	GuestAccessURL    string `json:"guest_access_url"`
}

func RoomFromHipchatv2(room hipchatv2.Room) (newRoom Room) {
	newRoom.ID = room.ID
	newRoom.JID = room.JID
	newRoom.Name = room.Name
	newRoom.Topic = room.Topic
	newRoom.Privacy = room.Privacy
	newRoom.IsArchived = room.IsArchived
	newRoom.IsGuestAccessible = room.IsGuestAccessible
	newRoom.GuestAccessURL = room.GuestAccessURL
	return
}

func CanonicalRoom(room string) string {
	if !strings.Contains(room, "@") {
		room += "@" + ConfDomain
	}
	return room
}

func BaseRoom(room string) string {
	if strings.Contains(room, "@") {
		return strings.Split(room, "@")[0]
	}
	return room
}
