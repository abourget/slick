package main

import (
	"github.com/golang/oauth2"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"github.com/GeertJohan/go.rice"
	"github.com/codegangsta/negroni"
	"html/template"
	"fmt"
)

type Webapp struct {
	config *WebappConfig
	store  *sessions.CookieStore
}

type WebappConfigSection struct {
	Webapp WebappConfig
}

type WebappConfig struct {
	ClientID          string `json:"client_id"`
	ClientSecret      string `json:"client_secret"`
	RestrictDomain    string `json:"restrict_domain"`
	SessionAuthKey    string `json:"session_auth_key"`
	SessionEncryptKey string `json:"session_encrypt_key"`
}

var notAuthenticatedTemplate = template.Must(template.New("").Parse(`
<html><body>
You have currently not given permissions to access your data. Please authenticate this app with the Google OAuth provider.
<form action="/authorize" method="POST"><input type="submit" value="Ok, authorize this app with my id"/></form>
</body></html>
`))

var userInfoTemplate = template.Must(template.New("").Parse(`
<html><body>
This app is now authenticated to access your Google user info.  Your details are:<br />
{{.}}<br />
Name: {{.Name}}<br />
Email: {{.Email}} Verified: {{.VerifiedEmail}}<br />
Hd: {{.Hd}}<br />
</body></html>
`))

func launchWebapp() {
	var conf WebappConfigSection
	bot.LoadConfig(&conf)

	web = &Webapp{
		config: &conf.Webapp,
		store:  sessions.NewCookieStore([]byte(conf.Webapp.SessionAuthKey), []byte(conf.Webapp.SessionEncryptKey)),
	}

	var err error
	oauthCfg, err = oauth2.NewConfig(
		&oauth2.Options{
			ClientID:     conf.Webapp.ClientID,
			ClientSecret: conf.Webapp.ClientSecret,
			RedirectURL:  "http://localhost:8080/oauth2callback",
			Scopes:       []string{"openid", "profile", "email", "https://www.googleapis.com/auth/userinfo.profile"},
		},
		"https://accounts.google.com/o/oauth2/auth",
		"https://accounts.google.com/o/oauth2/token",
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(rice.MustFindBox("static").HTTPBox())))
	mux.HandleFunc("/", handleRoot)

	n := negroni.Classic()
	n.UseHandler(context.ClearHandler(NewOAuthMiddleware(mux)))

	n.Run("localhost:8080")
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
