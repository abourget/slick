package tabularasa

import (
	"fmt"
	"sync"

	"github.com/abourget/ahipbot"
	"github.com/abourget/ahipbot/asana"
)

type TabulaRasa struct {
	bot         *ahipbot.Bot
	asanaClient *asana.Client
}

func init() {
	ahipbot.RegisterPlugin(func(bot *ahipbot.Bot) ahipbot.Plugin {

		var asanaConf struct {
			Asana struct {
				APIKey    string `json:"api_key"`
				Workspace string `json:"workspace"`
			}
		}

		bot.LoadConfig(&asanaConf)

		asanaClient := asana.NewClient(asanaConf.Asana.APIKey, asanaConf.Asana.Workspace)

		tabula := &TabulaRasa{
			bot:         bot,
			asanaClient: asanaClient,
		}

		return tabula
	})

}

func (tabula *TabulaRasa) Config() *ahipbot.PluginConfig {
	return &ahipbot.PluginConfig{}
}

func (tabula *TabulaRasa) Handle(bot *ahipbot.Bot, msg *ahipbot.BotMessage) {
	return
}

func (tabula *TabulaRasa) tabulaRasterize() {

	taskhose := make(chan asana.Task, 100)

	wg := sync.WaitGroup

	users, err := tabula.asanaClient.GetUsers()

	if err != nil {
		fmt.Println("anasa Client: ", err)
		return
	}

	go spinUpTaskWorker(taskhose)
	go spinUpTaskWorker(taskhose)

	for _, user := range users {
		wg.Add(1)
		go getFullTasksByAssignee(user, taskhose, wg)

	}

	wg.Wait()
	close(taskhose)

}

func (tabula *TabulaRasa) GetFullTasksByAssignee(user asana.User, taskhose chan asana.Task, wg *sync.WaitGroup) {

	defer wg.Done()

	tasks, err := asana.GetTasksByAssignee(user.Id)

	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		fulltask, err := asana.GetTaskById(task.Id)
		fmt.Println(fulltask.Name, fulltask.Completed)

		if err != nil {
			continue
		}

		taskhose <- fulltask
	}

}

func spinUpTaskWorker(taskhose chan asana.Task) {

	for task := range taskhose {
		if task.Completed {
			continue
		}
		fmt.Println("uncompleted task: ", task)
	}

}
