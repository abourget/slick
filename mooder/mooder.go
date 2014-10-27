package mooder

import (
	"math/rand"
	"time"

	"github.com/plotly/plotbot"
)

type Mooder struct {
	bot *plotbot.Bot
}

func init() {
	plotbot.RegisterPlugin(&Mooder{})
}

func (mooder *Mooder) InitChatPlugin(bot *plotbot.Bot) {
	mooder.bot = bot
	go mooder.SetupMoodChanger()
}

func (mooder *Mooder) SetupMoodChanger() {
	bot := mooder.bot
	for {
		time.Sleep(10 * time.Second)
		newMood := plotbot.Happy

		rand.Seed(time.Now().UTC().UnixNano())

		happyChances := rand.Int() % 10
		if happyChances > 6 {
			newMood = plotbot.Hyper
		}

		bot.Mood = newMood

		bot.SendToRoom(bot.Config.TeamRoom, bot.WithMood("I'm quite happy today.", "I can haz!! It's going to be a great one today!!"))

		select {
		case <-plotbot.AfterNextWeekdayTime(time.Monday, 12, 0):
		case <-plotbot.AfterNextWeekdayTime(time.Tuesday, 12, 0):
		case <-plotbot.AfterNextWeekdayTime(time.Wednesday, 12, 0):
		case <-plotbot.AfterNextWeekdayTime(time.Thursday, 12, 0):
		case <-plotbot.AfterNextWeekdayTime(time.Friday, 12, 0):
		}
	}

}
