package plotberry

import (
	"encoding/json"
	"fmt"
	"github.com/plotly/plotbot"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type PlotBerry struct {
	bot        *plotbot.Bot
	TotalUsers int
	pingTime   time.Duration
}

type TotalUsers struct {
	Plotberries int `json:"plotberries"`
}

func init() {
	plotbot.RegisterPlugin(&PlotBerry{})
}

func (plotberry *PlotBerry) InitChatPlugin(bot *plotbot.Bot) {

	plotberry.bot = bot
	plotberry.pingTime = 5 * time.Second
	plotberry.TotalUsers = 90000

	go plotberry.launchWatcher()

	bot.ListenFor(&plotbot.Conversation{
		HandlerFunc: plotberry.ChatHandler,
	})
}

func (plotberry *PlotBerry) ChatHandler(conv *plotbot.Conversation, msg *plotbot.Message) {
	if msg.MentionsMe && msg.Contains("how many users") {
		conv.Reply(msg, fmt.Sprintf("We got %d users!", plotberry.TotalUsers))
	}
	return
}

func (plotberry *PlotBerry) launchWatcher() {

	for {
		var data TotalUsers

		time.Sleep(plotberry.pingTime)

		resp, err := http.Get("https://plot.ly/v0/plotberries")
		defer resp.Body.Close()

		if err != nil {
			log.Fatalf("could not fetch: %v", err)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("could not parse plotberry response body")
			continue
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Fatalf("could not parse plotberry return json")
			continue
		}

		plotberry.TotalUsers = data.Plotberries
	}

}
