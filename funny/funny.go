package funny

import (
	"fmt"
	"time"

	"github.com/abourget/ahipbot"
	"github.com/abourget/ahipbot/blaster"
)

type Funny struct {
}

func init() {
	ahipbot.RegisterPlugin(func(bot *ahipbot.Bot) ahipbot.Plugin {
		return &Funny{}
	})

	ahipbot.RegisterStringList("forcePush", []string{
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

var config = &ahipbot.PluginConfig{
	EchoMessages: false,
	OnlyMentions: false,
}

func (funny *Funny) Config() *ahipbot.PluginConfig {
	return config
}

func (funny *Funny) Handle(bot *ahipbot.Bot, msg *ahipbot.BotMessage) {
	if msg.BotMentioned {
		if msg.ContainsAny([]string{"excitement", "exciting"}) {
			bot.Reply(msg, "http://static.fjcdn.com/gifs/Japanese+kids+spongebob+toys_0ad21b_3186721.gif")

		} else if msg.Contains("you're funny") {
			bot.Reply(msg, "/me blushes")
		}
	}

	if msg.ContainsAny([]string{"what is your problem", "what's your problem"}) {
		bot.Reply(msg, "http://media4.giphy.com/media/19hU0m3TJe6I/200w.gif")
		return

	} else if msg.Contains("force push") {
		url := ahipbot.RandomString("forcePush")
		bot.Reply(msg, url)
		return
	}

	if msg.Contains("blast") {
		//url := "https://plot.ly/__internal/ping"
		//url := "https://plot.ly/"
		url := "https://stage.plot.ly/__internal/ping"
		go func() {
			bot.Reply(msg, fmt.Sprintf("Blasting URL: %s for 60 seconds, with 5 workers", url))
			b := blaster.New(url)
			b.Start(5, time.Duration(60*time.Second))
			for rep := range b.Reply {
				bot.Reply(msg, rep)
			}
		}()
	}
}
