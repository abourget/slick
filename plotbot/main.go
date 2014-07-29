package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/abourget/ahipbot"

	_ "github.com/abourget/ahipbot/deployer"
	_ "github.com/abourget/ahipbot/funny"
	_ "github.com/abourget/ahipbot/healthy"
	_ "github.com/abourget/ahipbot/standup"
	_ "github.com/abourget/ahipbot/storm"
)

var configFile = flag.String("config", os.Getenv("HOME")+"/.hipbot", "config file")

func main() {
	flag.Parse()

	// TODO: most of this could and should go in "ahipbot"
	// we shouldn't know about what needs to be configured.. unless it's useful here..
	bot := ahipbot.NewHipbot(*configFile)
	bot.LoadBaseConfig()
	bot.SetupStorage()

	// Web related
	go ahipbot.LaunchWebapp(bot)

	ahipbot.LoadPlugins(bot)
	// Bot related
	for {
		log.Println("Connecting client...")
		err := bot.ConnectClient()
		if err != nil {
			log.Println("  `- Failed: ", err)
			time.Sleep(3 * time.Second)
			continue
		}

		disconnect := bot.SetupHandlers()

		select {
		case <-disconnect:
			log.Println("Disconnected...")
			time.Sleep(1 * time.Second)
			continue
		}
	}
}
