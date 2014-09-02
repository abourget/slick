package toxin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (toxin *Toxin) WebPluginSetup(router *mux.Router) {
	router.HandleFunc("/toxin/{id}", toxin.renderMeeting)
}

func (toxin *Toxin) renderMeeting(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var meeting *Meeting
	for _, meetingEl := range toxin.pastMeetings {
		if meetingEl.ID == id {
			meeting = meetingEl
		}
	}

	if meeting == nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	err := json.NewEncoder(w).Encode(meeting)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
	}
}
