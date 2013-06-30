package main

import (
	"github.com/tkawachi/hipchat"
	"math/rand"
	"strings"
	"time"
	"regexp"
)

type Wikipedia struct{}

const regs = [
	regexp.Compile(`wikipedia\s+(.+)`) ]

var r = rand.New(rand.NewSource(99))

func (wikipedia *Wikipedia) Handle(msg *hipchat.Message) *HandleReply {
	if strings.HasPrefix(msg.Body, "wikipedia ") {
		// TODO Access wikipedia to get contents
		time.Sleep(time.Duration(r.Intn(10)) * time.Second)
		return &HandleReply{
			To:      msg.From,
			Message: "http://jp.wikipedia.org Searching wikipedia..." + msg.Body,
		}
	}
	return nil
}
