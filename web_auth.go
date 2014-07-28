package ahipbot

import (
	"encoding/json"
	"fmt"
	"github.com/golang/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const oauthProfileInfoURL = "https://www.googleapis.com/oauth2/v1/userinfo?alt=json"

var oauthCfg *oauth2.Config

type OAuthMiddlware struct {
	handler http.Handler
}

func NewOAuthMiddleware(handler http.Handler) *OAuthMiddlware {
	return &OAuthMiddlware{handler: handler}
}

func (mw *OAuthMiddlware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/oauth2callback" {
		handleOAuth2Callback(w, r)
		return
	}

	_, err := checkAuth(r)
	if err != nil {
		if r.URL.Path == "/" {
			log.Println("Not logged in", err)
			url := oauthCfg.AuthCodeURL("")
			http.Redirect(w, r, url, http.StatusFound)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
		return
	}

	// Check if session exists, yield a 403 unless we're on the main page
	mw.handler.ServeHTTP(w, r)
}

func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	profile, err := doOAuth2Roundtrip(w, r)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		// Mark logged in
		sess := getSession(r)
		sess.Values["profile"] = profile
		sess.Save(r, w)

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func doOAuth2Roundtrip(w http.ResponseWriter, r *http.Request) (*GoogleUserProfile, error) {
	code := r.FormValue("code")

	t, err := oauthCfg.NewTransportWithCode(code)
	if err != nil {
		log.Println("OAuth2: ", err)
		return nil, fmt.Errorf("Error processing token.  Did you try to refresh ?")
	}

	//now get user data based on the Transport which has the token
	client := http.Client{Transport: t}
	resp, err := client.Get(oauthProfileInfoURL)
	if err != nil {
		log.Fatal("Fatal error After OAuth2: ", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Could not read data from Google APIs: %s", err)
	}

	var profile = GoogleUserProfile{}
	err = json.Unmarshal(body, &profile)
	if err != nil {
		return nil, fmt.Errorf("Could not read data from Google APIs: %s", err)
	}

	if domainCheck := web.config.RestrictDomain; domainCheck != "" {
		if profile.Hd != "" && domainCheck != profile.Hd {
			return nil, fmt.Errorf("Wrong hosted domain: %s", profile.Email)
		}
		domain := "@" + domainCheck
		if !strings.HasSuffix(profile.Email, domain) {
			return nil, fmt.Errorf("Wrong domain: %s", profile.Email)
		}
	}
	if profile.VerifiedEmail == false {
		return nil, fmt.Errorf("Email not verified: %s", profile.Email)
	}

	return &profile, nil
}

func checkAuth(r *http.Request) (*GoogleUserProfile, error) {
	sess := getSession(r)
	rawProfile, ok := sess.Values["profile"]
	if ok == false {
		return nil, fmt.Errorf("Not authenticated")
	}
	profile, ok := rawProfile.(*GoogleUserProfile)
	if ok == false {
		return nil, fmt.Errorf("Profile data unreadable")
	}
	return profile, nil
}
