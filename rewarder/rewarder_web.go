package rewarder

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/plotly/plotbot"
)

func (rew *Rewarder) InitWebPlugin(bot *plotbot.Bot, router *mux.Router) {
	router.HandleFunc("/rewarder/badges.json", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Badges []*Badge `json:"badges"`
		}{rew.Badges()}

		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
		}
	})
}
