package web

import "github.com/gorilla/mux"

type WebPlugin interface {
	WebPluginSetup(*mux.Router)
}
