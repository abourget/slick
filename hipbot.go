package main

import (
	"encoding/json"
	"flag"
	"github.com/tkawachi/hipchat"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
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
	stormMode  bool
}

func NewHipbot(configFile string) *Hipbot {
	bot := &Hipbot{}
	bot.replySink = make(chan *BotReply)
	bot.configFile = configFile
	bot.stormMode = false
	return bot
}

func (bot *Hipbot) Reply(msg *BotMessage, reply string) {
	log.Println("Replying:", reply)
	bot.replySink <- msg.Reply(reply)
}

func (bot *Hipbot) Storm(room string) {
	log.Println("STORMING!")
	gif := "http://8tracks.imgix.net/i/002/361/684/astronaut-3818.gif"
	// msg += "\n STORMED"
	if !strings.Contains(room, "@") {
		room = room + "@" + ConfDomain
	}
	reply := &BotReply{
		To: room,
		Message: gif,
	}
	bot.replySink <- reply

	msg := "Stormed!"
	reply = &BotReply{
		To: room,
		Message: msg,
	}
	bot.replySink <- reply
}

func (bot *Hipbot) connectClient() (err error) {
	bot.client, err = hipchat.NewClient(
		bot.config.Username, bot.config.Password, "bot")
	if err != nil {
		return
	}

	for _, room := range bot.config.Rooms {
		if !strings.Contains(room, "@") {
			room = room + "@" + ConfDomain
		}
		bot.client.Join(room, bot.config.Nickname)
	}

	return
}

func (bot *Hipbot) setupHandlers() chan bool {
	bot.client.Status("chat")
	disconnect := make(chan bool)
	go bot.client.KeepAlive()
	go bot.replyHandler(disconnect)
	go bot.messageHandler(disconnect)
	go bot.disconnectHandler(disconnect)
	log.Println("hipbot started")
	return disconnect
}

func (bot *Hipbot) loadBaseConfig() {
	if err := checkPermission(bot.configFile); err != nil {
		log.Fatal("ERROR Checking Permissions: ", err)
	}

	var config Config
	err := bot.LoadConfig(&config)
	if err != nil {
		log.Fatal("ERROR parsing config: ", err)
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
	plugins = append(plugins, NewDeployer(bot))
	bot.plugins = plugins
}

func (bot *Hipbot) replyHandler(disconnect chan bool) {
	for {
		reply := <-bot.replySink
		if reply != nil {
			log.Println("REPLYING", reply.To, reply.Message)
			bot.client.Say(reply.To, bot.config.Nickname, reply.Message)
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (bot *Hipbot) messageHandler(disconnect chan bool) {
	msgs := bot.client.Messages()
	for {
		msg := <-msgs
		botMsg := &BotMessage{Message: msg}
		log.Println("MESSAGE", msg)

		atMention := "@" + bot.config.Mention
		if strings.Contains(msg.Body, atMention) || strings.HasPrefix(msg.Body, bot.config.Mention) {
			botMsg.BotMentioned = true
			log.Printf("Message to me from %s: %s\n", msg.From, msg.Body)
		}

		for _, p := range bot.plugins {
			pluginConf := p.Config()

			fromMyself := strings.HasPrefix(botMsg.FromNick(), bot.config.Nickname)
			if !pluginConf.EchoMessages && fromMyself {
				log.Printf("no echo but I just messaged myself")
				continue
			}
			if pluginConf.OnlyMentions && !botMsg.BotMentioned {
				log.Printf("only mentions but not BotMentioned")
				continue
			}

			go func(p Plugin) { p.Handle(bot, botMsg) }(p)
		}
	}
}

func (bot *Hipbot) disconnectHandler(disconnect chan bool) {
	select {
	case <-disconnect:
		return
	}
	close(disconnect)
}
