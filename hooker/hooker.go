package hooker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/plotly/plotbot"
)

func init() {
	plotbot.RegisterPlugin(&Hooker{})
}

type Hooker struct {
	bot    *plotbot.Bot
	config HookerConfig
}

type HookerConfig struct {
	StripeSecret string `json:"stripe_secret"`
	GitHubSecret string `json:"github_secret"`
}

func (hooker *Hooker) InitWebPlugin(bot *plotbot.Bot, privRouter *mux.Router, pubRouter *mux.Router) {
	hooker.bot = bot

	var conf struct {
		Hooker HookerConfig
	}
	bot.LoadConfig(&conf)
	hooker.config = conf.Hooker

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

		// TODO: unmarshal the JSON, and check "hooker.config.GitHubSecret"

		bot.SendToRoom("123823_devops", fmt.Sprintf("/code Got a webhook from Github:\n%s", body))
	})

	stripeUrl := fmt.Sprintf("/public/stripehook/%s", hooker.config.StripeSecret)
	pubRouter.HandleFunc(stripeUrl, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not accepted", 405)
			return
		}

		bodyBytes, _ := ioutil.ReadAll(r.Body)

		var stripeEvent struct {
			Type    string
			Id      string
			Request string
		}
		err := json.Unmarshal(bodyBytes, &stripeEvent)

		if err != nil {
			log.Println("Hooker: unable to decode incoming JSON: ", err)
			return
		}

		if stripeEvent.Type == "customer.subscription.created" {
			bot.SendToRoom(bot.Config.TeamRoom, fmt.Sprintf("Hey! Someone just subscribed to Plotly! More details here: https://dashboard.stripe.com/logs/%s", stripeEvent.Request))
		}
	})

	privRouter.HandleFunc("/plugins/hooker.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not accepted", 405)
			return
		}

	})
}
