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

func (tabula *TabulaRasa) TabulaRasta() {

	taskhose := make(chan asana.Task, 100)

	wg := &sync.WaitGroup{}

	users, err := tabula.asanaClient.GetUsers()

	if err != nil {
		fmt.Println("anasa Client: ", err)
		return
	}

	go tabula.SpinUpTaskWorker(taskhose)
	go tabula.SpinUpTaskWorker(taskhose)
	go tabula.SpinUpTaskWorker(taskhose)
	go tabula.SpinUpTaskWorker(taskhose)

	for _, user := range users {
		wg.Add(1)
		fmt.Println(user)
		go tabula.GetFullTasksByAssignee(user, taskhose, wg)

	}

	wg.Wait()
	close(taskhose)

}

func (tabula *TabulaRasa) GetFullTasksByAssignee(user asana.User, taskhose chan asana.Task, wg *sync.WaitGroup) {

	defer wg.Done()

	tasks, err := tabula.asanaClient.GetTasksByAssignee(user)

	if err != nil {
		fmt.Println("Error acquiring task ids in GetFullTasksByAssignee", err)
		return
	}

	for _, task := range tasks {

		fulltask, err := tabula.asanaClient.GetTaskById(task.Id)

		if err != nil {
			fmt.Println("Error aquiring full task GetFullTasksByAssignee", err)
			continue
		}

		taskhose <- *fulltask
	}

}

func (tabula *TabulaRasa) SpinUpTaskWorker(taskhose chan asana.Task) {

	for task := range taskhose {
		if task.Completed {
			continue
		}
		updatedTask, err := tabula.asanaClient.UpdateTask("assignee=null", task)

		if err != nil {
			fmt.Println("Error updating task ", task)
		}

		fmt.Println("updated task:", updatedTask)
	}

}
