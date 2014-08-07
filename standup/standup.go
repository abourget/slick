package standup

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/abourget/ahipbot"
)

type Standup struct {
	bot *ahipbot.Bot

	// Map's Hipchat ID to UserData
	data *DataMap
}

type DataMap map[int64]*UserData

type UserData struct {
	Name     string
	Email    string
	PhotoURL string
	Yesterday  string
	Today      string
	Blocking   string
	LastUpdate time.Time
}

func (ud *UserData) FirstName() string {
	return strings.Split(ud.Name, " ")[0]
}

var config = &ahipbot.PluginConfig{
	EchoMessages: false,
	OnlyMentions: false,
}

func (standup *Standup) Config() *ahipbot.PluginConfig {
	return config
}

func init() {
	gob.Register(&UserData{})
	ahipbot.RegisterPlugin(func(bot *ahipbot.Bot) ahipbot.Plugin {
		dataMap := make(DataMap)
		standup := &Standup{bot: bot, data: &dataMap}
		standup.LoadData()
		return standup
	})
}

func (standup *Standup) Handle(bot *ahipbot.Bot, msg *ahipbot.BotMessage) {
	if strings.HasPrefix(msg.Body, "!yesterday") {
		standup.StoreLine(bot, msg, TYPE_YESTERDAY, msg.Body)

	} else if strings.HasPrefix(msg.Body, "!today") {
		standup.StoreLine(bot, msg, TYPE_TODAY, msg.Body)

	} else if strings.HasPrefix(msg.Body, "!blocking") {
		standup.StoreLine(bot, msg, TYPE_BLOCKING, msg.Body)

	} else if strings.HasPrefix(msg.Body, "!done?") {
		standup.ShowWhatsDone(bot, msg)

	}
}

type LineType int8

const (
	TYPE_YESTERDAY LineType = iota
	TYPE_TODAY
	TYPE_BLOCKING
)

func (standup *Standup) StoreLine(bot *ahipbot.Bot, msg *ahipbot.BotMessage, lineType LineType, line string) {
	user := bot.GetUser(msg.From)
	if user == nil {
		bot.Reply(msg, "Couldn't find your user profile.. have you just logged in? Wait a sec and try again.")
		return
	}

	dataMap := *standup.data
	userData, ok := dataMap[user.ID]
	if !ok {
		userData = &UserData{Email: user.Email, Name: user.Name, PhotoURL: user.PhotoURL}
		dataMap[user.ID] = userData
	}

	if lineType == TYPE_YESTERDAY {
		userData.Yesterday = line
	} else if lineType == TYPE_TODAY {
		userData.Today = line
	} else if lineType == TYPE_BLOCKING {
		userData.Blocking = line
	}

	userData.LastUpdate = time.Now().UTC()

	standup.FlushData()
}

func (standup *Standup) ShowWhatsDone(bot *ahipbot.Bot, msg *ahipbot.BotMessage) {
	user := bot.GetUser(msg.From)
	if user == nil {
		bot.Reply(msg, "Couldn't find your user profile.. have you just logged in? Wait a sec and try again.")
		return
	}

	dataMap := *standup.data
	userData, ok := dataMap[user.ID]
	if !ok {
		bot.Reply(msg, "No data for you buddy, ever!")
		return
	}

	bot.Reply(msg, fmt.Sprintf("Here's your stuff %s\n%s\n%s\n%s", userData.FirstName(), userData.Yesterday, userData.Today, userData.Blocking))
}

func (standup *Standup) LoadData() {
	bot := standup.bot
	redis := bot.RedisPool.Get()
	defer redis.Close()

	fixup := func() {
		dataMap := make(DataMap)
		standup.data = &dataMap
	}

	res, err := redis.Do("GET", "plotbot:standup")
	if err != nil {
		log.Println("Standup: Couldn't load data from redis. Using fresh data.")
		fixup()
	}

	asBytes, _ := res.([]byte)
	dec := gob.NewDecoder(bytes.NewBuffer(asBytes))
	err = dec.Decode(standup.data)
	if err != nil {
		log.Println("Standup: Unable to decode data from redis. Using fresh data.")
		fixup()
	}
}

func (standup *Standup) FlushData() {
	bot := standup.bot
	redis := bot.RedisPool.Get()
	defer redis.Close()

	buf := bytes.NewBuffer([]byte(""))
	enc := gob.NewEncoder(buf)
	enc.Encode(standup.data)

	_, err := redis.Do("SET", "plotbot:standup", buf.String())
	if err != nil {
		log.Println("ERROR: Couldn't redis FlushData()")
	}
}
