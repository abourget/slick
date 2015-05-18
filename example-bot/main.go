package main

import (
	"flag"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/abourget/slick"
	_ "github.com/abourget/slick/bugger"
	_ "github.com/abourget/slick/deployer"
	_ "github.com/abourget/slick/funny"
	_ "github.com/abourget/slick/healthy"
	_ "github.com/abourget/slick/hooker"
	_ "github.com/abourget/slick/mooder"
	_ "github.com/abourget/slick/plotberry"
	_ "github.com/abourget/slick/rewarder"
	_ "github.com/abourget/slick/standup"
	_ "github.com/abourget/slick/totw"
	_ "github.com/abourget/slick/web"
	_ "github.com/abourget/slick/webutils"
	_ "github.com/abourget/slick/wicked"
)

var configFile = flag.String("config", os.Getenv("HOME")+"/.slick.conf", "config file")

func main() {
	flag.Parse()

	bot := slick.New(*configFile)

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
