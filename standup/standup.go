package standup

import (
	"encoding/gob"

	"strings"
	"time"

	"github.com/plotly/plotbot"
)

type Standup struct {
	bot            *plotbot.Bot
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
	plotbot.RegisterPlugin(&Standup{})
}

func (standup *Standup) InitChatPlugin(bot *plotbot.Bot) {
	dataMap := make(DataMap)
	standup.bot = bot
	standup.data = &dataMap
	standup.sectionUpdates = make(chan sectionUpdate, 15)

	go standup.manageUpdatesInteraction()
	standup.LoadData()

	bot.ListenFor(&plotbot.Conversation{
		HandlerFunc: standup.ChatHandler,
	})
}

func (standup *Standup) ChatHandler(conv *plotbot.Conversation, msg *plotbot.Message) {
	res := sectionRegexp.FindAllStringSubmatchIndex(msg.Body, -1)
	if res != nil {
		for _, section := range extractSectionAndText(msg.Body, res) {
			standup.TriggerReminders(msg, section.name)
			standup.StoreLine(msg, section.name, section.text)
		}
	}
}
