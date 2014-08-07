package main

import (
	"flag"
	"os"

	"github.com/abourget/ahipbot"

	_ "github.com/abourget/ahipbot/deployer"
	_ "github.com/abourget/ahipbot/funny"
	_ "github.com/abourget/ahipbot/healthy"
	_ "github.com/abourget/ahipbot/standup"
	_ "github.com/abourget/ahipbot/storm"
)

var configFile = flag.String("config", os.Getenv("HOME")+"/.plotbot", "config file")

func main() {
	flag.Parse()

	bot := ahipbot.NewHipbot(*configFile)
	bot.Run()
}
