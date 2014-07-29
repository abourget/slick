package main

import (
	"time"
	"log"
	"strings"
)

type Storm struct {
	reminderCycle time.Duration
	config        *PluginConfig
}

type StormMode struct {
	on    bool
	link  string
}

var lastStorm = time.Now().UTC()
var stormTakerMsg = "IS THE STORM TAKER! \n" +
	"Go forth and summon the engineering powers of the team and transform " +
	"these requirements into tasks. If the requirements are incomplete or " +
	"confusing, it is your duty Storm Taker, yours alone, to remedy this. Good luck"
var stormMsg = "A storm is upon us! Who will step up and become the storm taker?!"

func NewStorm(bot *Hipbot) *Storm {
	storm := new(Storm)
	storm.reminderCycle = time.Second   //time (s) between storms
	storm.config = &PluginConfig{
		EchoMessages: true,
		OnlyMentions: false,
	}
	return storm
}

// Configuration
func (storm *Storm) Config() *PluginConfig {
	return storm.config
}

// Handler
func (storm *Storm) Handle(bot *Hipbot, msg *BotMessage) {

	//check for stormmode

	fromMyself := strings.HasPrefix(msg.FromNick(), bot.config.Nickname)
	room := "123823_devops"

	if msg.Contains("preparing storm") && fromMyself {
		// send first storms!
		bot.stormMode.on = true

		log.Println(bot.stormMode)

		sendStorm(room, RandomGIF("storm"))
		sendStorm(room, bot.stormMode.link)
		sendStorm(room, stormMsg)

	} else if bot.stormMode.on && !fromMyself {
		// storm taker!
		log.Println("Storm Taker!!!!!")

		stormTaker := msg.FromNick()
		stormTakerMsg = stormTaker + " " + stormTakerMsg
		bot.stormMode.on = false

		sendStorm(room, RandomGIF("herpderp"))
		sendStorm(room, stormTakerMsg)

	}
	// else if time.Since(lastStorm) > storm.reminderCycle {


	// 	//update laststorm
	// 	lastStorm = time.Now().UTC();
	// }

	return
}


func sendStorm(room string, message string) {

	if !strings.Contains(room, "@") {
		room = room + "@" + ConfDomain
	}

	reply := &BotReply{
		To: room,
		Message: message,
	}

	bot.replySink <- reply

	return
}
