package main

import (
	"github.com/GeertJohan/go.rice"
)

// This code is solely to trick "go-rice" into thinking we want those
// static assets.
func main() {
	rice.FindBox("static")
}
