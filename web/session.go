package session

import (
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

func GetSession(r *http.Request) *sessions.Session {
	sess, err := web.store.Get(r, "hipbot")
	if err != nil {
		log.Fatal(err)
	}

	return sess
}
