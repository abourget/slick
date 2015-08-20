package web

import (
	"encoding/json"
	"html/template"

	"github.com/nlopes/slack"
)

func userAsJavascript(user *slack.User) template.JS {
	jsonProfile, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		return template.JS(`{"error": "couldn't decode user"}`)
	}
	return template.JS(jsonProfile)
}
