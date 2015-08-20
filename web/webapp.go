package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/nlopes/slack"
	"github.com/abourget/slick"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var web *Webapp

type Webapp struct {
	config                *WebappConfig
	store                 *sessions.CookieStore
	bot                   *slick.Bot
	handler               *negroni.Negroni
	privateRouter         *mux.Router
	publicRouter          *mux.Router
	enabledPlugins        []string
	authMiddleware        func(http.Handler) http.Handler
	authenticatedUserFunc func(req *http.Request) (*slack.User, error)
}

type WebappConfig struct {
	Listen            string `json:"listen"`
	SessionAuthKey    string `json:"session_auth_key"`
	SessionEncryptKey string `json:"session_encrypt_key"`
}

func init() {
	slick.RegisterPlugin(&Webapp{})
}

func (webapp *Webapp) InitWebServer(bot *slick.Bot, enabledPlugins []string) {
	var conf struct {
		Webapp WebappConfig
	}
	bot.LoadConfig(&conf)

	webapp.bot = bot
	webapp.enabledPlugins = enabledPlugins
	webapp.config = &conf.Webapp
	webapp.store = sessions.NewCookieStore([]byte(conf.Webapp.SessionAuthKey), []byte(conf.Webapp.SessionEncryptKey))
	webapp.privateRouter = mux.NewRouter()
	webapp.publicRouter = mux.NewRouter()

	webapp.privateRouter.HandleFunc("/", webapp.handleRoot)

	web = webapp
}

func (webapp *Webapp) PrivateRouter() *mux.Router {
	return webapp.privateRouter
}
func (webapp *Webapp) PublicRouter() *mux.Router {
	return webapp.publicRouter
}

// SetAuthMiddleware should be called once by a WebServerAuth plugin, if any.
func (webapp *Webapp) SetAuthMiddleware(middleware func(http.Handler) http.Handler) {
	webapp.authMiddleware = middleware
}

// SetAuthenticatedUserFunc should be called once by a WebServerAuth plugin, if any.
func (webapp *Webapp) SetAuthenticatedUserFunc(f func(req *http.Request) (*slack.User, error)) {
	webapp.authenticatedUserFunc = f
}

func (webapp *Webapp) AuthenticatedUser(req *http.Request) (*slack.User, error) {
	if webapp.authenticatedUserFunc == nil {
		return nil, fmt.Errorf("No WebServerAuth plugin registered any AuthenticatedUser func call")
	}
	return webapp.authenticatedUserFunc(req)
}

func (webapp *Webapp) RunServer() {
	privMux := http.NewServeMux()
	privMux.Handle("/", webapp.PrivateRouter())

	pubMux := http.NewServeMux()
	pubMux.Handle("/public/", webapp.PublicRouter())
	if webapp.authMiddleware != nil {
		pubMux.Handle("/", webapp.authMiddleware(privMux))
	} else {
		pubMux.Handle("/", privMux)
	}

	webapp.handler = negroni.Classic()
	webapp.handler.UseHandler(context.ClearHandler(pubMux))

	webapp.handler.Run(webapp.config.Listen)
}

func (webapp *Webapp) GetSession(r *http.Request) *sessions.Session {
	sess, err := web.store.Get(r, "slick")
	if err != nil {
		log.Println("web/session: warn: unable to decode Session cookie: ", err)
	}

	return sess
}

func (webapp *Webapp) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	sess := webapp.GetSession(r)

	profile, err := webapp.AuthenticatedUser(r)

	fmt.Println("MAMA", profile, sess.Values, err)

	tpl, err := getRootTemplate()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	ctx := struct {
		CurrentUser, EnabledPlugins template.JS
	}{
		userAsJavascript(profile),
		webapp.getEnabledPluginsJS(),
	}
	tpl.Execute(w, ctx)
}

func getRootTemplate() (*template.Template, error) {
	return template.New("index.html").Parse(`
<html><head></head><body>
<script>
USER = {{.CurrentUser}};
ENABLED_PLUGINS = {{.EnabledPlugins}};
</script>
<h1>Welcome to your secure website</h1>
</body></html>
`)
}

func (webapp *Webapp) getEnabledPluginsJS() template.JS {
	out := make(map[string]bool)
	for _, pluginName := range webapp.enabledPlugins {
		out[pluginName] = true
	}

	jsonMap, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Println("Couldn't marshal EnabledPlugins list for rendering", err)
		return template.JS("{}")
	}
	return template.JS(jsonMap)
}
