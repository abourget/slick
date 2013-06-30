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

func main() {
	flag.Parse()
	config := loadConfig(*configFile)
	chatConfig := config.Hipchat
	log.Println(chatConfig)

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
	go client.KeepAlive()
	msgs := client.Messages()
	for {
		msg := <-msgs
		if msg.Delay == nil {
			if strings.HasPrefix(msg.Body, "wikipedia ") {
				log.Println("wikipedia, from:", msg.From)
				client.Say(msg.From, chatConfig.Nickname, "Searching wikipedia ...")
			}
		}
	}
}
