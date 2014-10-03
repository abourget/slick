package hooker

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/plotly/plotbot"
)

func init() {
	plotbot.RegisterPlugin(&Hooker{})
}

type Hooker struct {
	bot *plotbot.Bot
}

func (hooker *Hooker) InitWebPlugin(bot *plotbot.Bot, privRouter *mux.Router, pubRouter *mux.Router) {
	hooker.bot = bot

	pubRouter.HandleFunc("/public/updated_plotbot_repo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not accepted", 405)
			return
		}

		bodyBytes, err := ioutil.ReadAll(r.Body)

		body := ""
		if err != nil {
			body = "[Error reading body]"
		} else {
			body = string(bodyBytes)
		}

		bot.SendToRoom("123823_devops", fmt.Sprintf("/code Got a webhook from Github:\n%s", body))
	})

	privRouter.HandleFunc("/plugins/hooker.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not accepted", 405)
			return
		}

	})
}
