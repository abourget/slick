package webauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/abourget/slack"
	"github.com/abourget/slick"
	"golang.org/x/oauth2"
)

func init() {
	slick.RegisterPlugin(&OAuthPlugin{})
}

type OAuthPlugin struct {
	config    OAuthConfig
	webserver slick.WebServer
}

type OAuthConfig struct {
	RedirectURL  string `json:"oauth_base_url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (p *OAuthPlugin) InitWebServerAuth(bot *slick.Bot, webserver slick.WebServer) {
	p.webserver = webserver

	var config struct {
		WebAuthConfig OAuthConfig
	}
	bot.LoadConfig(&config)

	conf := config.WebAuthConfig
	webserver.SetAuthMiddleware(func(handler http.Handler) http.Handler {
		return &OAuthMiddleware{
			handler:   handler,
			plugin:    p,
			webserver: webserver,
			bot:       bot,
			oauthCfg: &oauth2.Config{
				ClientID:     conf.ClientID,
				ClientSecret: conf.ClientSecret,
				RedirectURL:  conf.RedirectURL + "/oauth2callback",
				Scopes:       []string{"identify"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://slack.com/oauth/authorize",
					TokenURL: "https://slack.com/api/oauth.access",
				},
			},
		}
	})
}

func (p *OAuthPlugin) AuthenticatedUser(r *http.Request) (*slack.User, error) {
	sess := p.webserver.GetSession(r)

	rawProfile, ok := sess.Values["profile"]
	if ok == false {
		return nil, fmt.Errorf("Not authenticated")
	}
	profile, ok := rawProfile.(*slack.User)
	if ok == false {
		return nil, fmt.Errorf("Profile data unreadable")
	}
	return profile, nil
}

type OAuthMiddleware struct {
	handler   http.Handler
	plugin    *OAuthPlugin
	webserver slick.WebServer
	oauthCfg  *oauth2.Config
	bot       *slick.Bot
}

func (mw *OAuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/oauth2callback" {
		mw.handleOAuth2Callback(w, r)
		return
	}

	_, err := mw.plugin.AuthenticatedUser(r)
	if err != nil {
		if r.URL.Path == "/" {
			log.Println("Not logged in", err)
			url := mw.oauthCfg.AuthCodeURL("", oauth2.SetAuthURLParam("team", mw.bot.Config.TeamDomain))
			http.Redirect(w, r, url, http.StatusFound)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
		return
	}

	// Check if session exists, yield a 403 unless we're on the main page
	mw.handler.ServeHTTP(w, r)
}

func (mw *OAuthMiddleware) handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	profile, err := mw.doOAuth2Roundtrip(w, r)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		// Mark logged in
		sess := mw.webserver.GetSession(r)
		sess.Values["profile"] = profile
		sess.Save(r, w)

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (mw *OAuthMiddleware) doOAuth2Roundtrip(w http.ResponseWriter, r *http.Request) (*slack.User, error) {
	code := r.FormValue("code")

	token, err := mw.oauthCfg.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Println("OAuth2: ", err)
		return nil, fmt.Errorf("Error processing token.  Did you try to refresh ?")
	}

	//now get user data based on the Transport which has the token
	client := mw.oauthCfg.Client(oauth2.NoContext, token)
	//mw.bot.api.
	resp, err := client.Get("https://www.googleapis.com/oauth2/v1/userinfo?alt=json")
	if err != nil {
		log.Fatal("Fatal error After OAuth2: ", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Could not read data from Google APIs: %s", err)
	}

	identity := struct {
		UserID string `json:"user_id"`
		URL    string
		OK     bool
	}{}
	err = json.Unmarshal(body, &identity)
	if err != nil {
		return nil, fmt.Errorf("Could not read data from Google APIs: %s", err)
	}
	// TODO; do something with userid, fetch from our internal list.

	if identity.URL != fmt.Sprintf("https://%s.slack.com") {
		return nil, fmt.Errorf("Authenticated for wrong domain: %s", identity.URL)
	}

	return &slack.User{}, nil
}
