package ahipbot

import (
	"encoding/gob"
	"html/template"
	"encoding/json"
	"log"
)

type GoogleUserProfile struct {
	Name          string `json:"name"`
	Hd            string `json:"hd"`
	Email         string `json:"email"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

func (up *GoogleUserProfile) AsJavascript() template.JS {
	jsonProfile, err := json.MarshalIndent(up, "", "  ")
	if err != nil {
		log.Fatal("Couldn't unmarshal Cookie with GoogleUserProfile in there", err)
	}
	return template.JS(jsonProfile)
}

func init() {
	gob.Register(&GoogleUserProfile{})

	// Backwards compatibility
	type UserProfile GoogleUserProfile
	gob.Register(&UserProfile{})
}
