package plotbot

type Mood int

const (
	Angry Mood = iota
	Happy
)

func (bot *Bot) WithMood(happy, angry string) string {
	if bot.Mood == Angry {
		return angry
	} else {
		return happy
	}
}
