package main

import (
	"encoding/json"
	"flag"
	"github.com/tkawachi/hipchat"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var configFile = flag.String("config", os.Getenv("HOME")+"/.hipbot", "config file")

const (
	ConfDomain = "conf.hipchat.com"
)

type Hipbot struct {
	configFile string
	config     HipchatConfig
	client     *hipchat.Client
	plugins    []Plugin
	replySink  chan *BotReply
}

func NewHipbot(configFile string) *Hipbot {
	bot := &Hipbot{}
	bot.replySink = make(chan *BotReply)
	bot.configFile = configFile
	return bot
}

func (bot *Hipbot) Reply(msg *BotMessage, reply string) {
	log.Println("Replying:", reply)
	bot.replySink <- msg.Reply(reply)
}

func (bot *Hipbot) connectClient() {
	var err error
	bot.client, err = hipchat.NewClient(
		bot.config.Username, bot.config.Password, "bot")
	if err != nil {
		log.Fatal(err)
	}
	for _, room := range bot.config.Rooms {
		if !strings.Contains(room, "@") {
			room = room + "@" + ConfDomain
		}
		bot.client.Join(room, bot.config.Nickname)
	}
}

func (bot *Hipbot) setupHandlers() {
	bot.client.Status("chat")
	go bot.client.KeepAlive()
	go bot.replyHandler()
	go bot.messageHandler()
	log.Println("hipbot started")
}

func (bot *Hipbot) loadBaseConfig() {
	if err := checkPermission(bot.configFile); err != nil {
		log.Fatal(err)
	}

	var config Config
	err := bot.LoadConfig(&config)
	if err != nil {
		log.Fatal(err)
	}

	bot.config = config.Hipchat
}

func (bot *Hipbot) LoadConfig(config interface{}) (err error) {
	content, err := ioutil.ReadFile(bot.configFile)
	if err != nil {
		log.Fatalln("ERROR reading config:", err)
		return
	}
	err = json.Unmarshal(content, &config)
	return
}

func (bot *Hipbot) registerPlugins() {
	plugins := make([]Plugin, 0)
	plugins = append(plugins, NewHealthy(bot))
	plugins = append(plugins, NewFunny(bot))
	bot.plugins = plugins
}

func (bot *Hipbot) replyHandler() {
	for {
		reply := <-bot.replySink
		if reply != nil {
			bot.client.Say(reply.To, bot.config.Nickname, reply.Message)
		}
	}
}

func (bot *Hipbot) messageHandler() {
	msgs := bot.client.Messages()
	for {
		msg := <-msgs
		botMsg := &BotMessage{Message: msg}

		atMention := "@" + bot.config.Mention
		if strings.Contains(msg.Body, atMention) || strings.HasPrefix(msg.Body, bot.config.Mention) {
			botMsg.BotMentioned = true
			log.Printf("Mentioned by %s: %s\n", msg.From, msg.Body)
		}

		for _, p := range bot.plugins {
			pluginConf := p.Config()

			fromMyself := strings.HasPrefix(botMsg.FromNick(), bot.config.Nickname)
			if !pluginConf.EchoMessages && fromMyself {
				continue
			}
			if !pluginConf.OnlyMentions && !botMsg.BotMentioned {
				continue
			}

			go func(p Plugin) { p.Handle(bot, botMsg) }(p)
		}
	}
}
