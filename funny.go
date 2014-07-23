package main

import (
	"math/rand"
	"time"
)

type Funny struct {
	config *PluginConfig
}

func NewFunny(bot *Hipbot) *Funny {
	return &Funny{config: &PluginConfig{
		EchoMessages: false,
		OnlyMentions: false,
	}}
}

// Configuration
func (funny *Funny) Config() *PluginConfig {
	return funny.config
}

var forcePushes = []string{
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
}

// Handler
func (funny *Funny) Handle(bot *Hipbot, msg *BotMessage) {
	if msg.BotMentioned {
		if msg.ContainsAny([]string{"excitement", "exciting"}) {
			bot.Reply(msg, "http://static.fjcdn.com/gifs/Japanese+kids+spongebob+toys_0ad21b_3186721.gif")
		}
		return
	}

	// Anywhere
	if msg.ContainsAny([]string{"what is your problem", "what's your problem"}) {
		bot.Reply(msg, "http://media4.giphy.com/media/19hU0m3TJe6I/200w.gif")
		return
	} else if msg.Contains("force push") {
		rand.Seed(time.Now().UTC().UnixNano())
		idx := rand.Int() % len(forcePushes)
		url := forcePushes[idx]
		bot.Reply(msg, url)
		return
	}

}
