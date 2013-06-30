package main

import (
	"flag"
	"github.com/tkawachi/hipchat"
	"log"
	"os"
	"strings"
)

var configFile = flag.String("config", os.Getenv("HOME")+"/.hipbot", "config file")

const (
	ConfDomain = "conf.hipchat.com"
)

type HandleReply struct {
	To      string
	Message string
}

type Plugin interface {
	Handle(*hipchat.Message) *HandleReply
}

var plugins = make([]Plugin, 0)

func registerPlugins() {
	plugins = append(plugins, new(Wikipedia))
}

func replyHandler(client *hipchat.Client, nickname string, respCh chan *HandleReply) {
	for {
		reply := <-respCh
		if reply != nil {
			client.Say(reply.To, nickname, reply.Message)
		}
	}
}

func messageHandler(client *hipchat.Client, plugins []Plugin, respCh chan *HandleReply) {
	msgs := client.Messages()
	for {
		msg := <-msgs
		for _, plugin := range plugins {
			go func() { respCh <- plugin.Handle(msg) }()
		}
	}
}

func main() {
	flag.Parse()
	config := loadConfig(*configFile)
	registerPlugins()
	chatConfig := config.Hipchat
	client, err := hipchat.NewClient(
		chatConfig.Username, chatConfig.Password, chatConfig.Resource)
	if err != nil {
		log.Fatal(err)
	}
	for _, room := range chatConfig.Rooms {
		if !strings.Contains(room, "@") {
			room = room + "@" + ConfDomain
		}
		client.Join(room, chatConfig.Nickname)
	}
	respCh := make(chan *HandleReply)
	go client.KeepAlive()
	go replyHandler(client, chatConfig.Nickname, respCh)
	go messageHandler(client, plugins, respCh)
	log.Println("hipbot started")
	select {}
}
