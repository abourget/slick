package main

import (
	"github.com/tkawachi/hipbot/plugin"
	"github.com/tkawachi/hipchat"
	"math/rand"
	"strings"
	"time"
)

type Wikipedia struct{}

var r = rand.New(rand.NewSource(99))

func (wikipedia *Wikipedia) Handle(msg *hipchat.Message) *plugin.HandleReply {
	if strings.HasPrefix(msg.Body, "wikipedia ") {
		// TODO Access wikipedia to get contents
		time.Sleep(time.Duration(r.Intn(10)) * time.Second)
		return &plugin.HandleReply{
			To:      msg.From,
			Message: "http://jp.wikipedia.org Searching wikipedia..." + msg.Body,
		}
	}
	return nil
}
