package web

import (
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

func GetSession(r *http.Request) *sessions.Session {
	sess, err := web.store.Get(r, "plotbot")
	if err != nil {
		log.Println("web/session: warn: unable to decode Session cookie: ", err)
	}

	return sess
}
