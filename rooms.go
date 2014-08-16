package ahipbot

import (
	"strings"

	"github.com/abourget/ahipbot/hipchatv2"
)

type Room struct {
	hipchatv2.Room
}

func canonicalRoom(room string) string {
	if !strings.Contains(room, "@") {
		room += "@conf.hipchat.com"
	}
	return room
}

func baseRoom(room string) string {
	if strings.Contains(room, "@") {
		return strings.Split(room, "@")[0]
	}
	return room
}
