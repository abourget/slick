package main

import (
	"github.com/gorilla/sessions"
	"net/http"
	"log"
)

func getSession(r *http.Request) *sessions.Session {
	sess, err := web.store.Get(r, "hipbot")
	if err != nil {
		log.Fatal(err)
	}

	return sess
}
