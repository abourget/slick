package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/GeertJohan/go.rice"
	"github.com/abourget/ahipbot"
	"github.com/codegangsta/negroni"
	"github.com/golang/oauth2"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var web *Webapp

type Webapp struct {
	config  *WebappConfig
	store   *sessions.CookieStore
	bot     *ahipbot.Bot
	handler *negroni.Negroni
}

type WebappConfig struct {
	Listen            string `json:"listen"`
	OAuthBaseURL      string `json:"oauth_base_url"`
	ClientID          string `json:"client_id"`
	ClientSecret      string `json:"client_secret"`
	RestrictDomain    string `json:"restrict_domain"`
	SessionAuthKey    string `json:"session_auth_key"`
	SessionEncryptKey string `json:"session_encrypt_key"`
}

func init() {
	ahipbot.RegisterWebHandler(func(bot *ahipbot.Bot, plugins []ahipbot.Plugin) ahipbot.WebHandler {
		var conf struct {
			Webapp WebappConfig
		}
		bot.LoadConfig(&conf)

		webapp := &Webapp{
			bot:    bot,
			config: &conf.Webapp,
			store:  sessions.NewCookieStore([]byte(conf.Webapp.SessionAuthKey), []byte(conf.Webapp.SessionEncryptKey)),
		}

		configureWebapp(&conf.Webapp)

		web = webapp

		rt := mux.NewRouter()
		rt.HandleFunc("/", handleRoot)

		for _, plugin := range plugins {
			webPlugin, ok := plugin.(WebPlugin)
			if !ok {
				continue
			}
			webPlugin.WebPluginSetup(rt)
		}

		mux := http.NewServeMux()
		mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(rice.MustFindBox("static").HTTPBox())))
		mux.Handle("/", rt)

		webapp.handler = negroni.Classic()
		webapp.handler.UseHandler(context.ClearHandler(NewOAuthMiddleware(mux)))
		return webapp
	})
}

func (webapp *Webapp) Run() {
	webapp.handler.Run(webapp.config.Listen)
}

// func LaunchWebapp(b *ahipbot.Bot) {

// 	rt.HandleFunc("/send_notif", handleNotif)
// 	rt.HandleFunc("/hipchat/users", handleGetUsers)
// 	rt.HandleFunc("/hipchat/rooms", handleGetRooms)

// 	n.Run("localhost:8080")
// }

func configureWebapp(conf *WebappConfig) {
	var err error
	oauthCfg, err = oauth2.NewConfig(
		&oauth2.Options{
			ClientID:     conf.ClientID,
			ClientSecret: conf.ClientSecret,
			RedirectURL:  conf.OAuthBaseURL + "/oauth2callback",
			Scopes:       []string{"openid", "profile", "email", "https://www.googleapis.com/auth/userinfo.profile"},
		},
		"https://accounts.google.com/o/oauth2/auth",
		"https://accounts.google.com/o/oauth2/token",
	)
	if err != nil {
		log.Fatal("oauth2.NewConfig error: ", err)
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
