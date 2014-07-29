package ahipbot

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"encoding/json"

	"github.com/GeertJohan/go.rice"
	"github.com/abourget/ahipbot/hipchatv2"
	"github.com/codegangsta/negroni"
	"github.com/golang/oauth2"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var web *Webapp

type Webapp struct {
	config *WebappConfig
	store  *sessions.CookieStore
}

type WebappConfig struct {
	ClientID          string `json:"client_id"`
	ClientSecret      string `json:"client_secret"`
	RestrictDomain    string `json:"restrict_domain"`
	SessionAuthKey    string `json:"session_auth_key"`
	SessionEncryptKey string `json:"session_encrypt_key"`
}

var bot *Hipbot

func LaunchWebapp(b *Hipbot) {
	bot = b
	var conf struct {
		Webapp WebappConfig
	}
	bot.LoadConfig(&conf)

	web = &Webapp{
		config: &conf.Webapp,
		store:  sessions.NewCookieStore([]byte(conf.Webapp.SessionAuthKey), []byte(conf.Webapp.SessionEncryptKey)),
	}

	configureWebapp(&conf.Webapp)

	rt := mux.NewRouter()
	rt.HandleFunc("/", handleRoot)
	rt.HandleFunc("/send_notif", handleNotif)
	rt.HandleFunc("/hipchat/users", handleGetUsers)
	rt.HandleFunc("/hipchat/rooms", handleGetRooms)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(rice.MustFindBox("static").HTTPBox())))
	mux.Handle("/", rt)

	n := negroni.Classic()
	n.UseHandler(context.ClearHandler(NewOAuthMiddleware(mux)))

	n.Run("localhost:8080")
}

func configureWebapp(conf *WebappConfig) {
	var err error
	oauthCfg, err = oauth2.NewConfig(
		&oauth2.Options{
			ClientID:     conf.ClientID,
			ClientSecret: conf.ClientSecret,
			RedirectURL:  "http://localhost:8080/oauth2callback",
			Scopes:       []string{"openid", "profile", "email", "https://www.googleapis.com/auth/userinfo.profile"},
		},
		"https://accounts.google.com/o/oauth2/auth",
		"https://accounts.google.com/o/oauth2/token",
	)
	if err != nil {
		log.Fatal("oauth error: ", err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	profile, _ := checkAuth(r)

	tpl, err := getRootTemplate()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	tpl.Execute(w, profile)
}

func getRootTemplate() (*template.Template, error) {
	box, err := rice.FindBox("static")
	if err != nil {
		return nil, fmt.Errorf("Error finding static assets: %s", err)
	}

	rawTpl, err := box.String("index.html")
	if err != nil {
		return nil, fmt.Errorf("Error loading index.html: %s", err)
	}

	tpl, err := template.New("index.html").Parse(rawTpl)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse index.html: %s", err)
	}

	return tpl, nil
}

// Send a notification through Hipchat
func handleNotif(w http.ResponseWriter, r *http.Request) {
	hipchatv2.SendNotification(bot.Config.HipchatApiToken, "DevOps", "gray", "text", "Hey that's great!", false)
}

func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := hipchatv2.GetUsers(bot.Config.HipchatApiToken)
	if err != nil {
		webReportError(w, "Error fetching users", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	out := struct {
		Users []hipchatv2.User `json:"users"`
	}{
		Users: users,
	}

	err = enc.Encode(out)
	if err != nil {
		webReportError(w, "Error encoding JSON", err)
		return
	}
	return
}

func handleGetRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := hipchatv2.GetRooms(bot.Config.HipchatApiToken)
	if err != nil {
		webReportError(w, "Error fetching rooms", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	out := struct {
		Rooms []hipchatv2.Room `json:"rooms"`
	}{
		Rooms: rooms,
	}

	err = enc.Encode(out)
	if err != nil {
		webReportError(w, "Error encoding JSON", err)
		return
	}
	return
}

func webReportError(w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("%s\n\n%s\n", msg, err)))
}
