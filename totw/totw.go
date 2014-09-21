package totw

import (
	"log"
	"strings"
	"time"

	"github.com/plotly/plotbot"
)

type Totw struct {
	bot *plotbot.Bot
}

func init() {
	plotbot.RegisterPlugin(&Totw{})
}

func (totw *Totw) InitChatPlugin(bot *plotbot.Bot) {
	plotbot.RegisterStringList("useless techs", []string{
		"http://i.minus.com/ib2bUNs2W1CI1V.gif",
		"http://media.giphy.com/media/anl0wydLNhKus/giphy.gif",
		"http://www.ptc.dcs.edu/Moody/comphistory/cavemanwriting.gif",
		"http://i.imgur.com/VbzhAbd.gif",
		"http://www.patrickcarrion.com/wp-content/uploads/2014/05/mowingdressgif.gif",
		"http://cdn.shopify.com/s/files/1/0243/7593/products/MKSB023_UselessMachine_Animation_grande.gif",
		"http://i.imgur.com/CRuLGek.gif",
		"http://i.imgur.com/EteBF9K.gif",
		"http://www.ohmagif.com/wp-content/uploads/2011/12/useless-invention.gif",
	})
	plotbot.RegisterStringList("tech adept", []string{
		"you're a real tech adept",
		"what an investigator",
		"such deep search!",
		"a real innovator you are",
		"way to go, I'm impressed",
		"hope it's better than my own code",
		"noted, but double check, you can never be certain",
		"I'll take a look into this one",
		"you're generous!",
		"hurray!",
	})

	totw.bot = bot

	go totw.ScheduleAlerts("123823_plotly", time.Thursday, 16, 0)
}

var config = &plotbot.ChatPluginConfig{
	EchoMessages: false,
	OnlyMentions: false,
}

func (totw *Totw) ChatConfig() *plotbot.ChatPluginConfig {
	return config
}

func (totw *Totw) ChatHandler(bot *plotbot.Bot, msg *plotbot.Message) {
	if strings.HasPrefix(msg.Body, "!totw") || strings.HasPrefix(msg.Body, "!techoftheweek"){
		bot.ReplyMention(msg, plotbot.RandomString("tech adept"))
	}
}

func (totw *Totw) ScheduleAlerts(room string, w time.Weekday, hour, min int) {
	for {
		next, when := plotbot.NextWeekdayTime(w, hour, min)
		log.Println("TOTW: Next occurrence: ", next)

		<-time.After(when)

		totw.bot.SendToRoom(room, plotbot.RandomString("useless techs"))
		totw.bot.SendToRoom(room, `Time for some Tech of the Week! What's your pick ?  Start your line with "!techoftheweek"`)
	}
}
