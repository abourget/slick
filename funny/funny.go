package funny


	"github.com/abourget/ahipbot"
type Funny struct {
}

func init() {
	ahipbot.RegisterPlugin(func(bot *ahipbot.Hipbot) ahipbot.Plugin {
		return &Funny{}
	})
}

var config = &ahipbot.PluginConfig{
	EchoMessages: false,
	OnlyMentions: false,
}

func (funny *Funny) Config() *ahipbot.PluginConfig {
	return config
}


func (funny *Funny) Handle(bot *ahipbot.Hipbot, msg *ahipbot.BotMessage) {

	if msg.BotMentioned {

		if msg.ContainsAny([]string{"excitement", "exciting"}) {
			bot.Reply(msg, "http://static.fjcdn.com/gifs/Japanese+kids+spongebob+toys_0ad21b_3186721.gif")

		} else if msg.Contains("you're funny") {
			bot.Reply(msg, "/me blushes")
		}
	}

	//bot.Reply(msg,"hello")

	if msg.ContainsAny([]string{"what is your problem", "what's your problem"}) {
		bot.Reply(msg, "http://media4.giphy.com/media/19hU0m3TJe6I/200w.gif")
		return

	} else if msg.Contains("force push") {
		url := RandomGIF("herpderp")
		bot.Reply(msg, url)
		return
	}
}
