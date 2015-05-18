package standup

import (
	"encoding/gob"

	"strings"
	"time"

	"github.com/abourget/slick"
)

type Standup struct {
	bot            *slick.Bot
	sectionUpdates chan sectionUpdate

	// Map's Hipchat ID to UserData
	data *DataMap
}

type DataMap map[int64]*UserData

type UserData struct {
	Name       string
	Email      string
	PhotoURL   string
	Yesterday  string
	Today      string
	Blocking   string
	LastUpdate time.Time
}

func (ud *UserData) FirstName() string {
	return strings.Split(ud.Name, " ")[0]
}

func init() {
	gob.Register(&UserData{})
	slick.RegisterPlugin(&Standup{})
}

func (standup *Standup) InitChatPlugin(bot *slick.Bot) {
	dataMap := make(DataMap)
	standup.bot = bot
	standup.data = &dataMap
	standup.sectionUpdates = make(chan sectionUpdate, 15)

	go standup.manageUpdatesInteraction()
	standup.LoadData()

	bot.ListenFor(&slick.Conversation{
		HandlerFunc: standup.ChatHandler,
	})
}

func (standup *Standup) ChatHandler(conv *slick.Conversation, msg *slick.Message) {
	res := sectionRegexp.FindAllStringSubmatchIndex(msg.Body, -1)
	if res != nil {
		for _, section := range extractSectionAndText(msg.Body, res) {
			standup.TriggerReminders(msg, section.name)
			standup.StoreLine(msg, section.name, section.text)
		}
	}
}
