package funny

import (
	"fmt"
	"strings"
	"time"

	"github.com/abourget/slick"
)

type Funny struct {
}

func init() {
	slick.RegisterPlugin(&Funny{})
}

func (funny *Funny) InitPlugin(bot *slick.Bot) {

	slick.RegisterStringList("forcePush", []string{
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

	slick.RegisterStringList("robot jokes", []string{
		"http://timmybeanbrain.files.wordpress.com/2012/05/05242012_02-01.jpg",
		"http://timmybeanbrain.files.wordpress.com/2012/05/05242012_01-01.jpg",
		"http://timmybeanbrain.files.wordpress.com/2012/05/05232012_01-01.jpg",
		"http://timmybeanbrain.files.wordpress.com/2012/05/05017012_01-01.jpg",
		"http://timmybeanbrain.files.wordpress.com/2012/07/07022012_04-01.jpg",
	})

	slick.RegisterStringList("dishes", []string{
		"http://stream1.gifsoup.com/view6/4703823/monkey-doing-dishes-o.gif",
		"http://s3-ec.buzzfed.com/static/enhanced/webdr06/2013/6/24/16/anigif_enhanced-buzz-9769-1372104764-13.gif",
		"http://i.imgur.com/WIL27Br.gif",
	})

	bot.Listen(&slick.Listener{
		MessageHandlerFunc: funny.ChatHandler,
	})
}

func (funny *Funny) ChatHandler(listen *slick.Listener, msg *slick.Message) {
	bot := listen.Bot

	if msg.Contains("mama") {
		listen.Bot.Listen(&slick.Listener{
			ListenDuration: time.Duration(10 * time.Second),
			MessageHandlerFunc: func(listen *slick.Listener, msg *slick.Message) {
				if strings.Contains(msg.Text, "papa") {
					msg.Reply("3s", "yo rocker").DeleteAfter("3s")
					msg.AddReaction("wink")
					go func() {
						time.Sleep(3 * time.Second)
						msg.AddReaction("beer")
						time.Sleep(1 * time.Second)
						msg.RemoveReaction("wink")
					}()
				}
			},
		})
	}

	if msg.MentionsMe {
		if msg.Contains("you're funny") {

			if bot.Mood == slick.Happy {
				msg.Reply("/me blushes")
			} else {
				msg.Reply("here's another one")
				msg.Reply(slick.RandomString("robot jokes"))
			}

		} else if msg.ContainsAny([]string{"dumb ass", "dumbass"}) {

			msg.Reply("don't say such things")

		} else if msg.ContainsAny([]string{"thanks", "thank you", "thx", "thnks"}) {
			msg.Reply(bot.WithMood("my pleasure", "any time, just ask, I'm here for you, ffiieeewww!get a life"))

		} else if msg.Contains("how are you") && msg.MentionsMe {
			msg.ReplyMention(bot.WithMood("good, and you ?", "I'm wild today!! wadabout you ?"))
			bot.Listen(&slick.Listener{
				ListenDuration: 60 * time.Second,
				FromUser:       msg.FromUser,
				FromChannel:    msg.FromChannel,
				MentionsMeOnly: true,
				MessageHandlerFunc: func(listen *slick.Listener, msg *slick.Message) {
					msg.ReplyMention(bot.WithMood("glad to hear it!", "zwweeeeeeeeet !"))
					listen.Close()
				},
				TimeoutFunc: func(listen *slick.Listener) {
					msg.ReplyMention("well, we can catch up later")
					listen.Close()
				},
			})
		}
	}

	if msg.ContainsAny([]string{"lot of excitement", "that's exciting", "how exciting", "much excitement"}) {

		msg.Reply("http://static.fjcdn.com/gifs/Japanese+kids+spongebob+toys_0ad21b_3186721.gif")

	} else if msg.ContainsAny([]string{"what is your problem", "what's your problem", "is there a problem", "which problem"}) {

		msg.Reply("http://media4.giphy.com/media/19hU0m3TJe6I/200w.gif")

	} else if msg.Contains("force push") {

		url := slick.RandomString("forcePush")
		msg.Reply(url)

	} else if msg.ContainsAny([]string{"there is a bug", "there's a bug"}) {

		msg.Reply("https://s3.amazonaws.com/pushbullet-uploads/ujy7DF0U8wm-9YYvLZkmSM8pMYcxCXXig8LjJORE9Xzt/The-life-of-a-coder.jpg")

	} else if msg.ContainsAny([]string{"oh yeah", "approved"}) {

		msg.Reply("https://i.chzbgr.com/maxW250/4496881920/h9C58F860.gif")

	} else if msg.Contains("ice cream") {

		msg.Reply("http://i.giphy.com/IGyLuFXIGSJj2.gif")
		msg.Reply("I love ice cream too")

	} else if msg.ContainsAny([]string{"lot of tension", "some tension", " tensed"}) {

		msg.Reply("http://thumbpress.com/wp-content/uploads/2014/01/funny-gif-meeting-strangers-girl-scared1.gif")
		msg.Reply("tensed, like that ?")

	} else if msg.Contains("quick fix") {

		msg.Reply("http://blog.pgi.com/wp-content/uploads/2013/02/jim-carey.gif")
		msg.Reply("make it real quick")

	} else if msg.ContainsAny([]string{"crack an egg", "crack something", "to crack"}) {

		msg.Reply("http://s3-ec.buzzfed.com/static/enhanced/webdr02/2012/11/8/18/anigif_enhanced-buzz-31656-1352415875-9.gif")
		msg.Reply("crack an egg, yeah")

	} else if msg.ContainsAny([]string{"i'm stuck", "I'm stuck", "we're stuck"}) {

		msg.Reply("http://media.giphy.com/media/RVlWx1msxnf7W/giphy.gif")
		msg.Reply("I'm stuck too!")

	} else if msg.ContainsAny([]string{"watching tv", "watch tv"}) {

		msg.Reply("http://i0.kym-cdn.com/photos/images/newsfeed/000/495/040/9ab.gif")
		msg.Reply("like that ?")

	} else if msg.ContainsAny([]string{"spider", "pee on", "inappropriate"}) {

		msg.Reply("https://i.chzbgr.com/maxW500/5626597120/hB2E11E61.gif")

	} else if msg.ContainsAny([]string{"a meeting", "an interview"}) {

		msg.Reply("like this one")
		msg.Reply("https://i.chzbgr.com/maxW500/6696664320/hFC69678C.gif")

	} else if msg.ContainsAny([]string{"it's odd", "it is odd", "that's odd", "that is odd", "it's awkward", "it is awkward", "that's awkward", "that is awkward"}) {

		term := "awkward"
		if msg.Contains("odd") {
			term = "odd"
		}
		msg.Reply(fmt.Sprintf("THAT's %s", term))
		msg.Reply("https://i.chzbgr.com/maxW500/8296294144/h7AC1001C.gif")

	} else if msg.Text == "ls" {

		msg.Reply("/code deploy/      Contributors-Guide/ image_server/     sheep_porn/     streambed/\nstreamhead/  README.md")

	} else if msg.ContainsAny([]string{"that's really cool", "that is really cool", "really happy"}) {

		msg.Reply("http://media.giphy.com/media/BlVnrxJgTGsUw/giphy.gif")

	} else if msg.ContainsAny([]string{"difficult problem", "hard problem"}) {

		msg.Reply("naming things, cache invalidation and off-by-1 errors are the two most difficult computer science problems")

	} else if msg.Contains("in theory") {

		msg.Reply("yeah, theory and practice perfectly match... in theory.")
	} else if msg.Contains("dishes") {

		msg.Reply(slick.RandomString("dishes"))

	} else if msg.Contains(" bean") {

		msg.Reply("http://media3.giphy.com/media/c35RMDO6luMaQ/500w.gif")

	} else if msg.Contains("steak") {

		msg.Reply("http://media.tumblr.com/tumblr_me6r52h1md1r6nno1.gif")

	} else if msg.ContainsAny([]string{"booze", "alcohol", "martini", " dog "}) {

		msg.Reply("http://media2.giphy.com/media/ZmJBjPdd44gXS/200w.gif")

	} else if msg.ContainsAny([]string{"internet", " tube "}) {

		msg.Reply("https://pbs.twimg.com/media/By0J3YHCcAA4UBo.jpg:large")

	}
}
