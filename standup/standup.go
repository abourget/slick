package standup

import (
	"encoding/json"
	"log"
	"time"

	"github.com/abourget/slick"
	"github.com/abourget/slick/util"
	"github.com/abourget/slack"
	levelutil "github.com/syndtr/goleveldb/leveldb/util"
)

type Standup struct {
	bot            *slick.Bot
	sectionUpdates chan sectionUpdate
}

const TODAY = 0
const WEEKAGO = -6 // [0,-6] == 7 days

func init() {
	slick.RegisterPlugin(&Standup{})
}

func (standup *Standup) InitChatPlugin(bot *slick.Bot) {
	standup.bot = bot
	standup.sectionUpdates = make(chan sectionUpdate, 15)

	go standup.manageUpdatesInteraction()

	bot.ListenFor(&slick.Conversation{
		HandlerFunc: standup.ChatHandler,
	})
}

func (standup *Standup) ChatHandler(conv *slick.Conversation, msg *slick.Message) {
	res := sectionRegexp.FindAllStringSubmatchIndex(msg.Text, -1)
	if res != nil {
		for _, section := range extractSectionAndText(msg.Text, res) {
			standup.TriggerReminders(msg, section.name)
			err := standup.StoreLine(msg, section.name, section.text)
			if err != nil {
				log.Println(err)
			}
		}
	} else if msg.MentionsMe && msg.Contains("standup report") {
		daysAgo := util.GetDaysFromQuery(msg.Text)
		smap, err := standup.getRange(getStandupDate(-daysAgo), getStandupDate(TODAY))
		if err != nil {
			log.Println(err)
			conv.Reply(msg, standup.bot.WithMood("Sorry, could not retrieve your report...",
				"I am the eggman and the walrus ate your report - Fzaow!"))
		} else {
			if msg.Contains(" my ") {
				conv.Reply(msg, "/quote "+smap.filterByEmail(msg.FromUser.Profile.Email).String())
			} else {
				conv.Reply(msg, "/quote "+smap.String())
			}
		}
	}
}

func (standup *Standup) getRange(from, to standupDate) (standupMap, error) {
	db := standup.bot.DB
	// Range is [Start, Limit) - ie, limit is not inclusive, so we bump date one next.
	srange := levelutil.Range{
		Start: standupKey{date: from}.key(),
		Limit: standupKey{date: to.next()}.key(),
	}
	iter := db.NewIterator(&srange, nil)

	// keep a map of users so we don't ask plotbot to grab users from the chatapp
	// that we have already loaded.
	seenUsers := make(map[string]standupUser)

	smap := make(standupMap)

	for iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		key := standupKeyFromBytes(iter.Key())
		email := key.email
		standupDate := key.date

		var user standupUser

		if val, ok := seenUsers[email]; ok {

			// grab an existing user from the lookup
			user = val
		} else {

			// initialize a new user from the messaging platform
			puser := standup.bot.GetUser(email)
			if puser == nil {
				puser = &slack.User{}
			}
			user = standupUser{puser, standupData{}}

			// store a copy of user (with blank data) inside map for later lookup
			seenUsers[email] = user
		}
		data := iter.Value()
		err := json.Unmarshal(data, &user.data)
		if err != nil {
			return smap, err
		}

		smap[standupDate] = append(smap[standupDate], user)
	}

	iter.Release()
	return smap, iter.Error()
}

func (standup *Standup) get(u standupUser, sd standupDate) (stand standupData, err error) {
	key := standupKey{sd, u.Profile.Email}.key()

	db := standup.bot.DB
	data, err := db.Get(key, nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &stand)
	return
}

func (standup *Standup) put(u standupUser, sd standupDate) (err error) {
	db := standup.bot.DB

	jdata, err := json.Marshal(u.data)
	if err != nil {
		return
	}

	key := standupKey{sd, u.Profile.Email}.key()
	err = db.Put(key, jdata, nil)
	return
}

func (standup *Standup) StoreLine(msg *slick.Message, section string, line string) error {

	standupDate := getStandupDate(TODAY)
	user := standupUser{msg.FromUser, standupData{}}
	data, err := standup.get(user, standupDate)
	if err != nil {
		log.Printf("Standup data for %s does not exist - using fresh\n", user.Name)
	}

	// update the userdata with data from the database
	user.data = data
	if section == "yesterday" {
		user.data.Yesterday = line
	} else if section == "today" {
		user.data.Today = line
	} else if section == "blocking" {
		user.data.Blocking = line
	}

	user.data.LastUpdate = time.Now().UTC()
	err = standup.put(user, standupDate)

	return err
}
