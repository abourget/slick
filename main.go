package main

import (
	"flag"
	"log"
	"time"
	"github.com/bpostlethwaite/ahipbot/asana"
)

var bot *Hipbot
var web *Webapp

func main() {
	flag.Parse()
	bot = NewHipbot(*configFile)

	// TODO: make this a goroutine to run the bot also
	go launchWebapp()


	bot.loadBaseConfig()
	bot.registerPlugins()

	asanaClient, err := asana.NewClient(
		"", "")
	if err != nil {
		log.Println("ASANA - Failed: ", err)
	}

	go StormWatch(asanaClient)

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
