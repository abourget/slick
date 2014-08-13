package ahipbot

import "strings"

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
