package funny

import (
	"fmt"
	"time"

	"github.com/plotly/plotbot"
	"github.com/plotly/plotbot/blaster"
)

type Funny struct {
}

func init() {
	plotbot.RegisterPlugin(&Funny{})
}

func (funny *Funny) InitChatPlugin(bot *plotbot.Bot) {

	plotbot.RegisterStringList("forcePush", []string{
		"http://www.gifcrap.com/g2data/albums/TV/Star%20Wars%20-%20Force%20Push%20-%20Goats%20fall%20over.gif",
		"http://i.imgur.com/ZvZR6Ff.jpg",
		"http://i3.kym-cdn.com/photos/images/original/000/014/538/5FCNWPLR2O3TKTTMGSGJIXFERQTAEY2K.gif",
		"http://i167.photobucket.com/albums/u123/KevinB550/FORCEPUSH/starwarsagain.gif",
		"http://i.imgur.com/dqSIv6j.gif",
		"http://www.gifcrap.com/g2data/albums/TV/Star%20Wars%20-%20Force%20Push%20-%20Gun%20breaks.gif",
		"http://media0.giphy.com/media/qeWa5wV5aeEHC/giphy.gif",
		"http://img40.imageshack.us/img40/2529/obiwan20is20a20jerk.gif",
		"http://img856.imageshack.us/img856/2364/obiwanforcemove.gif",
		"http://img526.imageshack.us/img526/4750/bc6.gif",
		"http://img825.imageshack.us/img825/6373/tumblrluaj77qaoa1qzrlhg.gif",
		"http://img543.imageshack.us/img543/6222/basketballdockingbay101.gif",
		"http://img687.imageshack.us/img687/5711/frap.gif",
		"http://img96.imageshack.us/img96/812/starpigdockingbay101.gif",
		"http://img2.wikia.nocookie.net/__cb20131117184206/halo/images/2/2a/Xt0rt3r.gif",
	})
}

var config = &plotbot.ChatPluginConfig{
	EchoMessages: false,
	OnlyMentions: false,
}

func (funny *Funny) ChatConfig() *plotbot.ChatPluginConfig {
	return config
}

func (funny *Funny) ChatHandler(bot *plotbot.Bot, msg *plotbot.Message) {
	if msg.BotMentioned {
		if msg.Contains("you're funny") {

			bot.Reply(msg, bot.WithMood("/me blushes", "buzz off"))

		} else if msg.ContainsAny([]string{"dumb ass", "dumbass"}) {

			bot.Reply(msg, bot.WithMood("don't say such things", "you stink"))

		} else if msg.Contains("blast") {
			url := "https://plot.ly/__internal/ping"
			//url := "https://plot.ly/"
			//url := "https://stage.plot.ly/__internal/ping"
			go func() {
				bot.Reply(msg, fmt.Sprintf("Blasting URL: %s for 60 seconds, with 2 workers", url))
				b := blaster.New(url)
				b.Start(2, time.Duration(60*time.Second))
				for rep := range b.Reply {
					bot.Reply(msg, rep)
				}
			}()
		} else if msg.ContainsAny([]string{"thanks", "thank you", "thx", "thnks"}) {
			bot.Reply(msg, bot.WithMood("my pleasure", "get a life"))

			if bot.Rewarder != nil {
				fmt.Println("Ok, in here")
				bot.Rewarder.LogEvent(msg.FromUser, "thanks", nil)
			}
		}
	}

	if msg.ContainsAny([]string{"lot of excitement", "that's exciting", "how exciting", "much excitement"}) {

		bot.Reply(msg, "http://static.fjcdn.com/gifs/Japanese+kids+spongebob+toys_0ad21b_3186721.gif")
		return

	} else if msg.ContainsAny([]string{"what is your problem", "what's your problem", "is there a problem", "which problem"}) {

		bot.Reply(msg, "http://media4.giphy.com/media/19hU0m3TJe6I/200w.gif")
		return

	} else if msg.Contains("force push") {

		url := plotbot.RandomString("forcePush")
		bot.Reply(msg, url)
		return

	} else if msg.ContainsAny([]string{"there is a bug", "there's a bug"}) {

		bot.Reply(msg, "https://s3.amazonaws.com/pushbullet-uploads/ujy7DF0U8wm-9YYvLZkmSM8pMYcxCXXig8LjJORE9Xzt/The-life-of-a-coder.jpg")
		return

	} else if msg.ContainsAny([]string{"oh yeah", "approved"}) {

		bot.Reply(msg, "https://i.chzbgr.com/maxW250/4496881920/h9C58F860.gif")
		return

	} else if msg.ContainsAny([]string{"spider", "pee on", "inappropriate"}) {

		bot.Reply(msg, "https://i.chzbgr.com/maxW500/5626597120/hB2E11E61.gif")
		return

	} else if msg.ContainsAny([]string{"a meeting", "an interview"}) {

		bot.Reply(msg, "like this one")
		bot.Reply(msg, "https://i.chzbgr.com/maxW500/6696664320/hFC69678C.gif")
		return

	} else if msg.ContainsAny([]string{"gotta go", "have to go", "uroclub"}) {

		bot.Reply(msg, "When you gotta go, you gotta go")
		bot.Reply(msg, "https://i.chzbgr.com/maxW250/7159139072/hB63619C4.gif")
		return

	} else if msg.ContainsAny([]string{"it's odd", "it is odd", "that's odd", "that is odd", "it's awkward", "it is awkward", "that's awkward", "that is awkward"}) {

		term := "awkward"
		if msg.Contains("odd") {
			term = "odd"
		}
		bot.Reply(msg, fmt.Sprintf("THAT's %s", term))
		bot.Reply(msg, "https://i.chzbgr.com/maxW500/8296294144/h7AC1001C.gif")
		return

	}

	if msg.Body == "ls" {
		bot.Reply(msg, "/code deploy/      Contributors-Guide/ image_server/     sheep_porn/     streambed/\nstreamhead/  README.md")
	}

	return
}
