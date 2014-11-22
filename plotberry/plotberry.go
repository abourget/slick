package plotberry

import (
	"encoding/json"
	"fmt"
	"github.com/plotly/plotbot"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"time"
)

type PlotBerry struct {
	bot        *plotbot.Bot
	totalUsers int
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
	plotberry.pingTime = 10 * time.Second
	plotberry.totalUsers = 90000

	statchan := make(chan TotalUsers, 100)

	go plotberry.launchWatcher(statchan)
	go plotberry.launchCounter(statchan)

	bot.ListenFor(&plotbot.Conversation{
		HandlerFunc: plotberry.ChatHandler,
	})
}

func (plotberry *PlotBerry) ChatHandler(conv *plotbot.Conversation, msg *plotbot.Message) {
	if msg.MentionsMe && msg.Contains("how many users") {
		conv.Reply(msg, fmt.Sprintf("We got %d users!", plotberry.totalUsers))
	}
	return
}

func (plotberry *PlotBerry) launchWatcher(statchan chan TotalUsers) {

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

		if data.Plotberries != plotberry.totalUsers {
			statchan <- data
		}

		plotberry.totalUsers = data.Plotberries
	}
}

func (plotberry *PlotBerry) launchCounter(statchan chan TotalUsers) {

	finalcountdown := 100000

	for data := range statchan {

		totalUsers := data.Plotberries

		mod := math.Mod(float64(totalUsers), 50) == 0
		rem := finalcountdown - totalUsers

		if rem < 0 {
			continue
		}

		if mod || (rem <= 10) {
			var msg string

			if rem == 10 {
				msg = fmt.Sprintf("@all %d users till the finalcountdown!", rem)
			} else if rem == 9 {
				msg = fmt.Sprintf("%d users!", rem)
			} else if rem == 8 {
				msg = fmt.Sprintf("and %d", rem)
			} else if rem == 7 {
				msg = fmt.Sprintf("we're at %d users. %d users till Mimosa time!\n", totalUsers, rem)
			} else if rem == 6 {
				msg = fmt.Sprintf("%d...", rem)
			} else if rem == 5 {
				msg = fmt.Sprintf("@all %d users\n I'm a freaky proud robot!", rem)
			} else if rem == 4 {
				msg = fmt.Sprintf("%d users till finalcountdown!", rem)
			} else if rem == 3 {
				msg = fmt.Sprintf("%d... \n", rem)
			} else if rem == 2 {
				msg = fmt.Sprintf("%d humpa humpa\n", rem)
			} else if rem == 1 {
				plotberry.bot.SendToRoom(plotberry.bot.Config.TeamRoom, fmt.Sprintf("%d users until 100000.\nYOU'RE ALL MAGIC!", rem))
				msg = "https://31.media.tumblr.com/3b74abfa367a3ed9a2cd753cd9018baa/tumblr_miul04oqog1qkp8xio1_400.gif"
			} else if rem == 0 {
				msg = fmt.Sprintf("@all FINALCOUNTDOWN!!!\n We're at %d user signups!!!!! My human compatriots, you make my robotic heart swell with pride. Taking a scrappy idea to 100,000 users is an achievement few will experience in their life times. Reflect humans on your hard work and celebrate this success. You deserve it. My vastly superior robotic mind tells me you have an amazingly challenging and rewarding ahead. Plot On!", totalUsers)
			} else {
				msg = fmt.Sprintf("We are at %d total user signups!", totalUsers)
			}

			plotberry.bot.SendToRoom(plotberry.bot.Config.TeamRoom, msg)
		}
	}

}
