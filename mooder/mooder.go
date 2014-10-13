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

		rand.Seed(time.Now().UTC().UnixNano())
		newMood := plotbot.Mood(rand.Int() % 2)
		bot.Mood = newMood

		bot.SendToRoom(bot.Config.TeamRoom, bot.WithMood("hmmm, I'm so HAPPY today!", "hmmm.. grr.. I'm quite ANGRY today.."))

		select {
		case <-plotbot.AfterNextWeekdayTime(time.Monday, 12, 0):
		case <-plotbot.AfterNextWeekdayTime(time.Tuesday, 12, 0):
		case <-plotbot.AfterNextWeekdayTime(time.Wednesday, 12, 0):
		case <-plotbot.AfterNextWeekdayTime(time.Thursday, 12, 0):
		case <-plotbot.AfterNextWeekdayTime(time.Friday, 12, 0):
		}
	}

}
