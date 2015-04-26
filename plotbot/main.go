package main

import (
	"flag"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/plotly/plotbot"

	_ "github.com/plotly/plotbot/rewarder"

	_ "github.com/plotly/plotbot/web"

	_ "github.com/plotly/plotbot/bugger"
	_ "github.com/plotly/plotbot/deployer"
	_ "github.com/plotly/plotbot/funny"
	_ "github.com/plotly/plotbot/healthy"
	_ "github.com/plotly/plotbot/hooker"
	_ "github.com/plotly/plotbot/mooder"
	_ "github.com/plotly/plotbot/plotberry"
	_ "github.com/plotly/plotbot/standup"
	_ "github.com/plotly/plotbot/totw"
	_ "github.com/plotly/plotbot/webutils"
	_ "github.com/plotly/plotbot/wicked"
)

var configFile = flag.String("config", os.Getenv("HOME")+"/.plotbot", "config file")

func main() {
	flag.Parse()

	bot := plotbot.NewHipbot(*configFile)

	var serverConf struct {
		Server struct {
			Pidfile string `json:"pid_file"`
		}
	}

	bot.LoadConfig(&serverConf)
	pid := os.Getpid()
	pidb := []byte(strconv.Itoa(pid))
	ioutil.WriteFile(serverConf.Server.Pidfile, pidb, 0755)

	bot.Run()
}
