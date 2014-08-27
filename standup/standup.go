package standup

import (
	"encoding/gob"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/abourget/ahipbot"
)

type Standup struct {
	bot            *ahipbot.Bot
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
		standup := &Standup{
			bot:            bot,
			data:           &dataMap,
			sectionUpdates: make(chan sectionUpdate, 15),
		}
		go standup.manageUpdatesInteraction()
		standup.LoadData()
		return standup
	})
}

func (standup *Standup) Handle(bot *ahipbot.Bot, msg *ahipbot.BotMessage) {
	res := sectionRegexp.FindAllStringSubmatchIndex(msg.Body, -1)
	if res != nil {
		for _, section := range extractSectionAndText(msg.Body, res) {
			standup.TriggerReminders(msg, section.name)
			standup.StoreLine(msg, section.name, section.text)
		}
	}
}

var sectionRegexp = regexp.MustCompile(`(?mi)^!(yesterday|today|blocking)`)

type sectionMatch struct {
	name string
	text string
}

// extractSectionAndText returns the "today", "this is what I did today" section from the result of "FindAllStringSubmatchindex" call.
func extractSectionAndText(input string, res [][]int) []sectionMatch {
	out := make([]sectionMatch, 0, 3)

	for i := 0; i < len(res); i++ {
		el := res[i]

		section := input[el[2]:el[3]] // (2,3) is second group's (start,end)
		strings.ToLower(section)

		var endFullText = len(input)
		if (i + 1) < len(res) {
			endFullText = res[i+1][0]
		}
		fullText := input[el[1]:endFullText]
		fullText = strings.TrimSpace(fullText)

		out = append(out, sectionMatch{section, fullText})
	}

	return out
}

func (standup *Standup) StoreLine(msg *ahipbot.BotMessage, section string, line string) {
	dataMap := *standup.data
	user := standup.bot.GetUser(msg.From)
	userData, ok := dataMap[user.ID]
	if !ok {
		userData = &UserData{Email: user.Email, Name: user.Name, PhotoURL: user.PhotoURL}
		dataMap[user.ID] = userData
	}

	if section == "yesterday" {
		userData.Yesterday = line
	} else if section == "today" {
		userData.Today = line
	} else if section == "blocking" {
		userData.Blocking = line
	}

	userData.LastUpdate = time.Now().UTC()

	standup.FlushData()
}

func (standup *Standup) TriggerReminders(msg *ahipbot.BotMessage, section string) {
	standup.sectionUpdates <- sectionUpdate{section, msg}
}

//
// Reminder to complete all sections and reception confirmation message
//

func (standup *Standup) manageUpdatesInteraction() {
	remindCh := make(chan *ahipbot.BotMessage)
	resetCh := make(chan *ahipbot.BotMessage)

	for {
		select {
		case update := <-standup.sectionUpdates:
			userEmail := update.msg.FromUser.Email
			progress := userProgressMap[userEmail]
			if progress == nil {
				progress = &userProgress{
					sectionsDone: make(map[string]bool),
					cancelTimer:  make(chan bool),
				}
				userProgressMap[userEmail] = progress
				progress.sectionsDone[update.section] = true
				go progress.waitAndCheckProgress(update.msg, remindCh)
				go progress.waitForReset(update.msg, resetCh)
			} else {
				close(progress.cancelTimer)

				progress.sectionsDone[update.section] = true
				numDone := len(progress.sectionsDone)
				if numDone == 3 {
					standup.bot.ReplyMention(update.msg, "got it!")
					delete(userProgressMap, update.msg.FromUser.Email)
				} else {
					progress.cancelTimer = make(chan bool)
					go progress.waitAndCheckProgress(update.msg, remindCh)
				}
			}

		case msg := <-resetCh:
			userEmail := msg.FromUser.Email
			progress := userProgressMap[userEmail]
			if progress != nil {
				close(progress.cancelTimer)
			}
			delete(userProgressMap, userEmail)

		case msg := <-remindCh:
			// Do the reminding for that user
			userEmail := msg.FromUser.Email
			userProgress := userProgressMap[userEmail]
			if userProgress == nil {
				continue
			}

			remains := make([]string, 0, 3)
			if userProgress.sectionsDone["today"] == false {
				remains = append(remains, "today")
			}
			if userProgress.sectionsDone["yesterday"] == false {
				remains = append(remains, "yesterday")
			}
			if userProgress.sectionsDone["blocking"] == false {
				remains = append(remains, "blocking stuff")
			}

			remain := strings.Join(remains, " or ")

			if remain != "" {
				standup.bot.ReplyMention(msg, fmt.Sprintf("what about %s ?", remain))
			}
		}
	}
}

type sectionUpdate struct {
	section string
	msg     *ahipbot.BotMessage
}

var userProgressMap = make(map[string]*userProgress)

type userProgress struct {
	sectionsDone map[string]bool
	cancelTimer  chan bool
}

func (up *userProgress) waitAndCheckProgress(msg *ahipbot.BotMessage, remindCh chan *ahipbot.BotMessage) {
	select {
	case <-time.After(30 * time.Second):
		remindCh <- msg
	case <-up.cancelTimer:
		return
	}
}

// waitForReset waits a couple of minutes and stops listening to that user altogether.  We want to poke the user once or twice if he's slow.. but not eternally.
func (up *userProgress) waitForReset(msg *ahipbot.BotMessage, resetCh chan *ahipbot.BotMessage) {
	<-time.After(3 * time.Minute)
	resetCh <- msg
}
