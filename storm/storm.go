package storm

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/abourget/ahipbot"
	"github.com/abourget/ahipbot/asana"
)

type Storm struct {
	bot               *ahipbot.Hipbot
	config            *StormConfig
	timeBetweenStorms time.Duration
	stormActive       bool
	stormLink         string
	asanaClient       *asana.Client
	triggerPolling    chan bool
}

type StormConfig struct {
	AsanaAPIKey    string `json:"asana_api_key"`
	AsanaWorkspace string `json:"asana_workspace"`
	HipchatRoom    string `json:"hipchat_room"`
	StormTagId     string `json:"storm_tag_id"`
}

func init() {
	ahipbot.RegisterPlugin(func(bot *ahipbot.Hipbot) ahipbot.Plugin {
		var conf struct {
			Storm StormConfig
		}
		bot.LoadConfig(&conf)

		asana := asana.NewClient(conf.Storm.AsanaAPIKey, conf.Storm.AsanaWorkspace)

		storm := &Storm{
			bot:               bot,
			config:            &conf.Storm,
			timeBetweenStorms: 60 * time.Second,
			asanaClient:       asana,
			triggerPolling:    make(chan bool, 10),
		}

		ahipbot.RegisterStringList("storm", []string{
			"http://static.tumblr.com/ikqttte/OlElnumnn/f9cb7_tumblr_lkfd09xr2y1qfuje9o1_500.gif",
			"http://media.giphy.com/media/8cdBgACkApvt6/giphy.gif",
			"http://25.media.tumblr.com/tumblr_luucaug87A1qluhjfo1_500.gif",
			"http://cdn.mdjunction.com/components/com_joomlaboard/uploaded/images/storm.gif",
			"http://www.churchhousecollection.com/resources/animated-jesus-calms-storm.gif",
			"http://i251.photobucket.com/albums/gg307/angellovernumberone/HEATHERS%20%20MIXED%20WATER%20ANIMATIONS/LightningStorm02.gif",
			"http://i.imgur.com/IF1QM.gif",
			"http://wac.450f.edgecastcdn.net/80450F/screencrush.com/files/2013/04/x-men-storm.gif",
			"http://i.imgur.com/SNLbnO8.gif?1",
		})

		go storm.launchWatcher()

		return storm
	})
}

type StormMode struct {
	on   bool
	link string
}

var lastStorm = time.Now().UTC()
var stormTakerMsg = "IS THE STORM TAKER! \n" +
	"Go forth and summon the engineering powers of the team and transform " +
	"these requirements into tasks. If the requirements are incomplete or " +
	"confusing, it is your duty Storm Taker, yours alone, to remedy this. Good luck"

// Configuration
var config = &ahipbot.PluginConfig{
	EchoMessages: false,
	OnlyMentions: false,
}

func (storm *Storm) Config() *ahipbot.PluginConfig {
	return config
}

// Handler
func (storm *Storm) Handle(bot *ahipbot.Hipbot, msg *ahipbot.BotMessage) {

	//check for stormmode

	fromMyself := strings.HasPrefix(msg.FromNick(), bot.Config.Nickname)
	room := storm.config.HipchatRoom

	if msg.Contains("preparing storm") && fromMyself {
		// send first storms!
		storm.stormActive = true

		log.Println(storm.stormActive)

	} else if msg.BotMentioned && msg.Contains("stormy day") {

		if storm.stormActive {
			bot.Reply(msg, "We're in the middle of a storm, can't you feel it ?")
		} else {
			bot.Reply(msg, "You'll know it soon enough")
			storm.triggerPolling <- true
		}

	} else if storm.stormActive && !fromMyself {
		log.Println("Storm Taker!!!!!")

		stormTaker := msg.FromNick()
		stormTakerMsg = stormTaker + " " + stormTakerMsg
		storm.stormActive = false

		bot.SendToRoom(room, ahipbot.RandomString("forcePush"))
		bot.SendToRoom(room, stormTakerMsg)

	} else if storm.stormActive && msg.Contains("ENOUGH") {
		storm.stormActive = false
		bot.Reply(msg, "ok, ok !")
	}
	// else if time.Since(lastStorm) > storm.timeBetweenStorms {

	// 	//update laststorm
	// 	lastStorm = time.Now().UTC();
	// }

	return
}

func (storm *Storm) StormNotifWhatever() {
	log.Println("STORMING!")
	storm.bot.SendToRoom(storm.config.HipchatRoom, "http://8tracks.imgix.net/i/002/361/684/astronaut-3818.gif")
}

func (storm *Storm) launchWatcher() {
	for {
		timeout := time.After(60 * time.Second)
		select {
		case <-timeout:
		case <-storm.triggerPolling:
		}

		storm.pollAsana()
	}
}

func (storm *Storm) pollAsana() {
	stormTagId := storm.config.StormTagId

	if storm.stormActive {
		time.Sleep(5 * time.Second)
		return
	}

	stormedTasks := readTasks()

	tasks, err := storm.asanaClient.GetTasksByTag(stormTagId)
	if err != nil {
		log.Println("Storm: Error fetching tasks by tag: ", err)
	}

	for _, task := range tasks {
		taskId := strconv.FormatInt(task.Id, 10)
		taskAlreadyStormed := stringInSlice(taskId, stormedTasks)

		if !taskAlreadyStormed {
			log.Println("NEW STORM TAG DETECTED")
			storm.startStorm(task)
			return
		}
	}
}

func (storm *Storm) startStorm(task asana.Task) {
	room := storm.config.HipchatRoom
	taskId := strconv.FormatInt(task.Id, 10)

	storm.stormLink = asanaLink + taskId
	storm.stormActive = true
	bot := storm.bot
	bot.SendToRoom(room, "/me sees a storm getting near")
	go func() {
		time.Sleep(5 * time.Second)
		bot.SendToRoom(room, "Ahh here it is!")
		bot.SendToRoom(room, ahipbot.RandomString("storm"))
		bot.SendToRoom(room, fmt.Sprintf("Take it at this address: %s", storm.stormLink))
	}()
	go func() {
		time.Sleep(10 * time.Second)
		bot.SendToRoom(room, "A storm is upon us! Who will step up and become the storm taker?!")
	}()

	writeTask(taskId)
}
