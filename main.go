package main

import (
	"flag"
	"log"
)

func main() {
	flag.Parse()
	bot := NewHipbot(*configFile)
	bot.loadBaseConfig()
	bot.registerPlugins()
	for {
		log.Println("Connecting client...")
		bot.connectClient()
		disconnect := bot.setupHandlers()
		select {
		case <-disconnect:
			log.Println("Disconnected...")
			continue
		}
	}
}
