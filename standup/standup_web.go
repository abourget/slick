package standup

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/plotly/plotbot"
)

func (standup *Standup) InitWebPlugin(bot *plotbot.Bot, privRouter *mux.Router, pubRouter *mux.Router) {
	privRouter.HandleFunc("/plugins/standup.json", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Users []*UserData
		}{
			Users: make([]*UserData, 0),
		}
		for _, value := range *standup.data {
			data.Users = append(data.Users, value)
		}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			webReportError(w, "Error encoding data", err)
		}
	})
}

func webReportError(w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("%s\n\n%s\n", msg, err)))
}
