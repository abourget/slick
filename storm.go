package main

import (
"math/rand"
"time"
)

type Storm struct {
	reminderCycle time.Duration 
	config *PluginConfig
}

var lastStorm =  time.Now().UTC(); 

func NewStorm(bot *Hipbot) *Storm {
	storm := new(Storm)
	storm.reminderCycle = 5*time.Minute//time (s) between storms 
	storm.config = &PluginConfig{
		EchoMessages: false,
		OnlyMentions: true,
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
	if bot.stormMode{

		// check to see if someone killed the storm 
		killed := (msg.BotMentioned && msg.ContainsAny([]string{"got it", "ENOUGH!"})

		if (killed){		

			log.Println("STORM KILLED!")
			gif := "http://38.media.tumblr.com/6aa84eb94b29716d630162a4f737a73d/tumblr_mxnrmbwf7c1s7jx17o1_400.gif"

			if !strings.Contains(room, "@") {
				room = room + "@" + ConfDomain
			}

			reply := &BotReply{
				To: room,
				Message: gif,
			}

			bot.replySink <- reply

			msg := "STORM KILLED!"

			reply = &BotReply{
				To: room,
				Message: msg,
			}

			bot.replySink <- reply	

		}else{

					//check to see if enough time has passed since the last storm 
		if time.Since(lastStorm) > storm.reminderCycle
		{

			log.Println("STORMING!")
			gif := "http://8tracks.imgix.net/i/002/361/684/astronaut-3818.gif"

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

		//update laststorm 
		lastStorm = time.Now().UTC(); 

		}

	}
}


