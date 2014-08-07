package standup

import (
	"fmt"
	"strings"
)
import "log"
import "encoding/gob"
import "bytes"
import "github.com/abourget/ahipbot"
import "time"

type Standup struct {
	bot *ahipbot.Bot

	// Map's Hipchat ID to UserData
	data *DataMap
}

type DataMap map[int64]*UserData

type UserData struct {
	Name     string
	Email    string
	Triplets []*Triplet
}

func (ud *UserData) FirstName() string {
	return strings.Split(ud.Name, " ")[0]
}

type Triplet struct {
	Yesterday string
	Today     string
	Blocking  string
	Created   time.Time
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
	gob.Register(&Triplet{})
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

	} else if strings.HasPrefix(msg.Body, "!flush") {
		standup.FlushData()

	} else if strings.HasPrefix(msg.Body, "!load") {
		standup.LoadData()
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
		userData = &UserData{Email: user.Email, Name: user.Name}
		dataMap[user.ID] = userData
	}

	// Is the latest Triplet good enough ?
	var t *Triplet
	if len(userData.Triplets) == 0 {
		t = &Triplet{Created: time.Now().UTC()}
		userData.Triplets = append(userData.Triplets, t)
	} else {
		t = userData.Triplets[len(userData.Triplets)-1]
	}

	if lineType == TYPE_YESTERDAY {
		t.Yesterday = line
	} else if lineType == TYPE_TODAY {
		t.Today = line
	} else if lineType == TYPE_BLOCKING {
		t.Blocking = line
	}
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

	if len(userData.Triplets) == 0 {
		// Normally would always have one triplet if userData exists for this user!
		log.Println("ERRROR: No Triplets in UserData structure!")
		bot.Reply(msg, "Hmm. you hit some unreachable code. You're lucky.")
		return
	}

	lastTriplet := userData.Triplets[len(userData.Triplets)-1]

	bot.Reply(msg, fmt.Sprintf("Here's your stuff %s\n%s\n%s\n%s", userData.FirstName(), lastTriplet.Yesterday, lastTriplet.Today, lastTriplet.Blocking))
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
		log.Println("ERROR: standup: Couldn't load data from redis")
		fixup()
	}

	asBytes, _ := res.([]byte)
	dec := gob.NewDecoder(bytes.NewBuffer(asBytes))
	err = dec.Decode(standup.data)
	if err != nil {
		log.Println("ERROR: standup: Unable to decode LoadData() data from redis")
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
