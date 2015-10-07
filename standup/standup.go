package standup

import "github.com/abourget/slick"

type Standup struct {
	bot            *slick.Bot
	sectionUpdates chan sectionUpdate
}

const TODAY = 0
const WEEKAGO = -6 // [0,-6] == 7 days

func init() {
	slick.RegisterPlugin(&Standup{})
}

func (standup *Standup) InitPlugin(bot *slick.Bot) {
	standup.bot = bot
	standup.sectionUpdates = make(chan sectionUpdate, 15)

	go standup.manageUpdatesInteraction()

	bot.Listen(&slick.Listener{
		MessageHandlerFunc: standup.ChatHandler,
	})
}

func (standup *Standup) ChatHandler(listen *slick.Listener, msg *slick.Message) {
	res := sectionRegexp.FindAllStringSubmatchIndex(msg.Text, -1)
	if res != nil {
		for _, section := range extractSectionAndText(msg.Text, res) {
			standup.TriggerReminders(msg, section.name)
			// err := standup.StoreLine(msg, section.name, section.text)
			// if err != nil {
			// 	log.Println(err)
			// }
		}
	}
}
