package plotbot

type Mood int

const (
	Happy Mood = iota
	Hyper
)

func (bot *Bot) WithMood(happy, hyper string) string {
	if bot.Mood == Happy {
		return happy
	} else {
		return hyper
	}
}
