package main

import (
	"flag"
	"os"

	"github.com/abourget/ahipbot"

	_ "github.com/abourget/ahipbot/web"

	_ "github.com/abourget/ahipbot/deployer"
	_ "github.com/abourget/ahipbot/toxin"
	_ "github.com/abourget/ahipbot/funny"
	_ "github.com/abourget/ahipbot/healthy"
	_ "github.com/abourget/ahipbot/standup"
	_ "github.com/abourget/ahipbot/storm"
	_ "github.com/abourget/ahipbot/tabularasa"
	_ "github.com/abourget/ahipbot/webutils"
)

var configFile = flag.String("config", os.Getenv("HOME")+"/.plotbot", "config file")

func main() {
	flag.Parse()

	bot := ahipbot.NewHipbot(*configFile)
	bot.Run()
}
