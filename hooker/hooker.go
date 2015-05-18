package hooker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/abourget/slick"
)

func init() {
	slick.RegisterPlugin(&Hooker{})
}

type Hooker struct {
	bot    *slick.Bot
	config HookerConfig
}

type HookerConfig struct {
	StripeSecret string `json:"stripe_secret"`
	GitHubSecret string `json:"github_secret"`
}

type MonitAlert struct {
	Host    string `json:"host"`
	Date    string `json:"date"`
	Service string `json:"service"`
	Alert   string `json:"alert"`
}

func (hooker *Hooker) InitWebPlugin(bot *slick.Bot, privRouter *mux.Router, pubRouter *mux.Router) {
	hooker.bot = bot

	var conf struct {
		Hooker HookerConfig
	}
	bot.LoadConfig(&conf)
	hooker.config = conf.Hooker

	pubRouter.HandleFunc("/public/updated_slick_repo", hooker.updatedSlickRepo)

	stripeUrl := fmt.Sprintf("/public/stripehook/%s", hooker.config.StripeSecret)
	pubRouter.HandleFunc(stripeUrl, hooker.onPayingUser)

	pubRouter.HandleFunc("/public/monit", hooker.onMonit)

	privRouter.HandleFunc("/plugins/hooker.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not accepted", 405)
			return
		}

	})
}

func (hooker *Hooker) updatedSlickRepo(w http.ResponseWriter, r *http.Request) {
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

	text := fmt.Sprintf("/code Got a webhook from Github:\n%s", body)
	fmt.Println("TEST: ", text)
	//bot.SendToRoom("123823_devops", )
}

func (hooker *Hooker) onPayingUser(w http.ResponseWriter, r *http.Request) {
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
		hooker.bot.SendToRoom(hooker.bot.Config.TeamRoom,
			fmt.Sprintf("Hey! Someone just subscribed to Plotly! More details here: https://dashboard.stripe.com/logs/%s",
				stripeEvent.Request))
	}
}

func (hooker *Hooker) onMonit(w http.ResponseWriter, r *http.Request) {

	var alert MonitAlert

	if r.Method != "POST" {
		http.Error(w, "Method not accepted", 405)
		return
	}

	defer r.Body.Close()

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Hooker: unable to read incoming Zapier request: ", err)
		return
	}

	err = json.Unmarshal(bodyBytes, &alert)

	if err != nil {
		log.Println("Hooker: unable to decode incoming Alert JSON: ", err)
		return
	}

	fmt.Println("TEST: ", alert)

	//bot.SendToRoom("123823_devops", )
}
