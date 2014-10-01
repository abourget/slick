package wicked

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/plotly/plotbot"
)

func (wicked *Wicked) InitWebPlugin(bot *plotbot.Bot, router *mux.Router) {
	router.HandleFunc("/wicked/{id}.json", wicked.renderMeetingJson)
	router.HandleFunc("/wicked/{id}.html", wicked.renderMeetingHtml)
}

func (wicked *Wicked) renderMeetingJson(w http.ResponseWriter, r *http.Request) {
	meeting := wicked.webGetMeeting(r)

	if meeting == nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	err := json.NewEncoder(w).Encode(meeting)

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
	}
}

func (wicked *Wicked) renderMeetingHtml(w http.ResponseWriter, r *http.Request) {
	meeting := wicked.webGetMeeting(r)

	if meeting == nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	tmpl, err := template.New("index").Funcs(template.FuncMap{
		"idify": func(input time.Time) string {
			inputStr := input.String()
			inputStr = strings.Replace(inputStr, " ", "-", -1)
			return strings.Replace(inputStr, ":", "-", -1)
		},
	}).Parse(`<!DOCTYPE html>
<html>
<head>
  <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
</head>
<body>
  <h1>Meeting W{{.ID}}</h1>
  <h2>Goal: {{.Goal}}</h2>
  <p>In room: {{.Room}}</p>
  <p>Created by: {{.CreatedBy.Fullname}}</p>
  <p>
    Started at: {{.StartTime}}
    {{if .EndTime.IsZero}}
    - NOT ENDED
    {{else}}
    - ended: {{.EndTime}}
    {{end}}
  </p>
  <p>Participants:
    <ul>
      {{range .Participants}}
      <li>
        <img src="{{.PhotoURL}}" style="max-width: 64px;" />
        {{.Fullname}}
      </li>
      {{end}}
    </ul>
  </p>


  <h3>Propositions / Decisions</h3>
  {{if .Decisions}}
    {{range .Decisions}}
    <div>
      <a id="decision-{{.Timestamp | idify}}" />
      <h4>{{if .IsProposition}}Proposition:{{else}}Decision:{{end}}</h4>
      {{.Text}}
      <p style="font-size: 90%;">proposed by: {{.AddedBy.Fullname}}, at {{.Timestamp}}</p>
      {{if .Plusplus}}
      <p>Vouched for by: {{range .Plusplus}}{{.From.Fullname}}, {{end}}</p>
      {{end}}
    </div>
    {{end}}
  {{else}}
    <p>Nothing proposed</p>
  {{end}}


  <h3>References</h3>
  {{if .Refs}}
    {{range .Refs}}
    <div>
      <a id="ref-{{.Timestamp | idify}}" />
      <p>
        {{if .URL}}
          <a href="{{.URL}}" target="_blank">{{.URL}}</a>
        {{end}}
        {{.Text}}
      </p>
      <p style="font-size: 90%;">proposed by: <a href="#msg-{{.Timestamp | idify}}">{{.AddedBy.Fullname}}, at {{.Timestamp}}</a></p>
    </div>
    {{end}}
  {{else}}
    <p>No references</p>
  {{end}}


  <h3>Logs</h3>
  {{if .Logs}}
    {{range .Logs}}
    <div>
      <a id="msg-{{.Timestamp | idify}}" />
      <strong title="{{.Timestamp}}">{{.From.Fullname}}</strong> {{.Text}}<br />
    </div>
    {{end}}
  {{else}}
    <p>No messages</p>
  {{end}}

</body>
</html>


`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = tmpl.Execute(w, meeting)

	if err != nil {
		log.Println("Wicked Web Error: ", err)
		http.Error(w, http.StatusText(500), 500)
	}
}

func (wicked *Wicked) webGetMeeting(r *http.Request) *Meeting {
	params := mux.Vars(r)
	id := params["id"]

	for _, meetingEl := range wicked.pastMeetings {
		if meetingEl.ID == id {
			return meetingEl
		}
	}

	return nil
}
