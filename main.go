package main

import (
	"flag"
)

func main() {
	flag.Parse()
	bot := NewHipbot(*configFile)
	bot.loadBaseConfig()
	bot.registerPlugins()
	bot.connectClient()
	bot.setupHandlers()
	select {}
}
