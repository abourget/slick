package webauth

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"

	"github.com/abourget/slack"
	"github.com/abourget/slick"
	"golang.org/x/oauth2"
)

func init() {
	slick.RegisterPlugin(&OAuthPlugin{})
	gob.Register(&slack.User{})
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
	webserver.SetAuthenticatedUserFunc(p.AuthenticatedUser)
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
			url := mw.oauthCfg.AuthCodeURL("", oauth2.SetAuthURLParam("team", mw.bot.Config.TeamID))
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
		err := sess.Save(r, w)
		if err != nil {
			fmt.Println("Error saving cookie:", err)
			w.Write([]byte(err.Error()))
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (mw *OAuthMiddleware) doOAuth2Roundtrip(w http.ResponseWriter, r *http.Request) (*slack.User, error) {
	code := r.FormValue("code")

	token, err := mw.oauthCfg.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Println("OAuth2: ", err)
		return nil, fmt.Errorf("Error processing token.")
	}

	client := slack.New(token.AccessToken)

	resp, err := client.AuthTest()
	if err != nil {
		return nil, fmt.Errorf("User unauthenticated: %s", err)
	}

	expectedURL := fmt.Sprintf("https://%s.slack.com/", mw.bot.Config.TeamDomain)
	if resp.URL != expectedURL {
		return nil, fmt.Errorf("Authenticated for wrong domain: %q != %q", resp.URL, expectedURL)
	}

	return mw.bot.GetUser(resp.UserId), nil
}
