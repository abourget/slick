package main

import (
	"encoding/gob"
	"html/template"
	"encoding/json"
	"log"
)

type UserProfile struct {
	Name          string `json:"name"`
	Hd            string `json:"hd"`
	Email         string `json:"email"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

func (up *UserProfile) AsJavascript() template.JS {
	jsonProfile, err := json.MarshalIndent(up, "", "  ")
	if err != nil {
		log.Fatal("Couldn't unmarshal Cookie with UserProfile in there", err)
	}
	return template.JS(jsonProfile)
}

func init() {
	gob.Register(&UserProfile{})
}
