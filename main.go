package main

import (
	"flag"
	"log"
	"time"
)

var bot *Hipbot
var web *Webapp

func main() {
	flag.Parse()
	bot = NewHipbot(*configFile)

	// TODO: make this a goroutine to run the bot also
	launchWebapp()

	bot.loadBaseConfig()
	bot.registerPlugins()

	for {
		log.Println("Connecting client...")
		err := bot.connectClient()
		if err != nil {
			log.Println("  `- Failed: ", err)
			time.Sleep(3 * time.Second)
			continue
		}

		disconnect := bot.setupHandlers()

		select {
		case <-disconnect:
			log.Println("Disconnected...")
			time.Sleep(1 * time.Second)
			continue
		}
	}
}
