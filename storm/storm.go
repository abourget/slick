package storm

import (
	"bytes"
	"html/template"
	"log"
	"strconv"
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
	CalmedTagId    string `json:"calmed_tag_id"`
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
	if msg.BotMentioned && msg.Contains("stormy day") {

		if storm.stormActive {
			bot.Reply(msg, "We're in the middle of a storm, can't you feel it ?")
		} else {
			bot.Reply(msg, "You'll know it soon enough")
			storm.triggerPolling <- true
		}

	} else if storm.stormActive && msg.Contains("ENOUGH") {
		storm.stormActive = false
		bot.Reply(msg, "ok, ok !")
	}

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
	if storm.stormActive {
		time.Sleep(5 * time.Second)
		return
	}

	stormTasks, err := storm.asanaClient.GetTasksByTag(storm.config.StormTagId)
	if err != nil {
		log.Println("Storm: Error fetching tasks by tag: ", err)
		return
	}
	calmedTasks, err := storm.asanaClient.GetTasksByTag(storm.config.CalmedTagId)
	if err != nil {
		log.Println("Storm: Error fetching tasks by tag: ", err)
		return
	}

	newStorms := tasksDifference(stormTasks, calmedTasks)

	if len(newStorms) > 0 {
		log.Println("Storm: New Storm!")
		fullTask, err := storm.asanaClient.GetTaskById(newStorms[0].Id)
		if err != nil {
			log.Println("Storm: Error fetching the Task by ID", err)
			return
		}
		storm.startStorm(fullTask)
	}
}

// TODO: what's that anyway ? move to config ?
var asanaLink = "https://app.asana.com/0/7221799638526/"

const DEBUG = true

type tplData map[string]interface{}

var stormTpl = template.Must(template.New("stormTpl").Parse(`
<p>
  <img class="image" src="{{.Image}}">
</p>
<p><b>Take it here</b>: <a href="{{.StormLink}}">{{.Task.Name}}</a> by <i>{{.Task.CreatedBy.Name}}</i></p>
`))

func (storm *Storm) startStorm(task *asana.Task) {
	room := storm.config.HipchatRoom
	taskId := strconv.FormatInt(task.Id, 10)

	storm.stormLink = asanaLink + taskId
	storm.stormActive = true
	bot := storm.bot
	bot.SendToRoom(room, "/me sees a storm approaching")
	go func() {
		wait := time.Duration(5)
		if DEBUG {
			wait = 0
		}
		time.Sleep(wait * time.Second)

		bot.SendToRoom(room, "@all Who will step up and take it on ?")
		img := ahipbot.RandomString("storm")
		data := tplData{
			"StormLink": storm.stormLink,
			"Task": task,
			"Image": img,
		}
		buf := bytes.NewBuffer([]byte(""))
		stormTpl.Execute(buf, data)
		res, err := bot.Notify(room, "gray", "html", buf.String(), true)
		if err != nil {
			log.Printf("ERROR: Storm: Notifying about the Storm failed: %s %#v %#v\n", err, res.Error, res.Result)
		}
	}()
	go storm.watchForTaker(task)
	go storm.watchForCalm(task)
}



var takerTpl = template.Must(template.New("takerTpl").Parse(`
<p><img src="{{.ForcePush}}"></p>
<p>We have a Taker ! It's:</p>
{{if .User.Photo.Image128}}
  <p><img src="{{.User.Photo.Image128}}"></p>
{{end}}
<p>{{.User.Name}}</p>
`))

func (storm *Storm) watchForTaker(task *asana.Task) {
	firstRun := true
	seenStories := make(map[int64]bool)
	for {
		if storm.stormActive == false {
			log.Println("Storm: Stopping Taker watch, apparently the storm has stopped!")
			return
		}

		stories, err := storm.asanaClient.GetTaskStories(task.Id)
		if err != nil {
			log.Println("ERROR: Storm: couldn't fetch stories: ", err)
			return
		}

		for _, story := range stories {
			if firstRun {
				seenStories[story.Id] = true
				continue
			}

			_, ok := seenStories[story.Id]
			if !ok {
				// New entry! Who's that ?!
				user, err := storm.asanaClient.GetUser(story.CreatedBy.Id)
				if err != nil {
					log.Println("ERROR: Storm: couldn't get the User that was reported by Asana: ", err)
					return
				}

				buf := bytes.NewBuffer([]byte(""))
				data := tplData{
					"User": user,
					"ForcePush": ahipbot.RandomString("forcePush"),
				}
				takerTpl.Execute(buf, data)
				storm.bot.Notify(storm.config.HipchatRoom, "green", "html", buf.String(), false)
				return
			}
		}

		firstRun = false
		time.Sleep(5 * time.Second)
	}
}

func (storm *Storm) watchForCalm(originalTask *asana.Task) {
	room := storm.config.HipchatRoom
	for {
		time.Sleep(15 * time.Second)
		tags, err := storm.asanaClient.GetTagsOnTask(originalTask.Id)
		if err != nil {
			log.Println("ERROR: Storm: watchForCalm(): Couldn't get tags for task")
			storm.bot.SendToRoom(room, "folks, the Storm is over, Asana is freaking out")
			storm.stormActive = false
			return
		}

		for _, tag := range tags {
			if storm.config.CalmedTagId == tag.StringId() {
				storm.bot.SendToRoom(room, "ok, folks, the Storm has been Calmed!")
				storm.stormActive = false
			}
		}

	}
}

// taskDifference returns Storm-tagged tasks without the Calmed-tagged tasks
func tasksDifference(storm []asana.Task, calmed []asana.Task) (out []asana.Task) {
	seen := make(map[int64]bool)
	for _, task := range calmed {
		seen[task.Id] = true
	}

	for _, task := range storm {
		_, ok := seen[task.Id]
		if !ok {
			out = append(out, task)
		}
	}

	return out
}
