package slick

import "github.com/gorilla/mux"

//
// Bot plugins
//

type Plugin interface{}

type PluginInitializer interface {
	InitPlugin(*Bot)
}

type WebServer interface {
	InitWebServer(*Bot, []string)
	ServeWebRequests()
	PrivateRouter() *mux.Router
	PublicRouter() *mux.Router
}

// WebPlugin initializes plugins with a `Bot` instance, a `privateRouter` and a `publicRouter`. All URLs handled by the `publicRouter` must start with `/public/`.
type WebPlugin interface {
	InitWebPlugin(*Bot, *mux.Router, *mux.Router)
}

var registeredPlugins = make([]Plugin, 0)

func RegisterPlugin(plugin Plugin) {
	registeredPlugins = append(registeredPlugins, plugin)
}

func initChatPlugins(bot *Bot) {
	for _, plugin := range registeredPlugins {
		chatPlugin, ok := plugin.(PluginInitializer)
		if ok {
			chatPlugin.InitPlugin(bot)
		}
	}
}

func initWebServer(bot *Bot, enabledPlugins []string) {
	for _, plugin := range registeredPlugins {
		webServer, ok := plugin.(WebServer)
		if ok {
			webServer.InitWebServer(bot, enabledPlugins)
			bot.WebServer = webServer
			return
		}
	}
}

func initWebPlugins(bot *Bot) {
	if bot.WebServer == nil {
		return
	}

	for _, plugin := range registeredPlugins {
		webPlugin, ok := plugin.(WebPlugin)
		if ok {
			webPlugin.InitWebPlugin(bot, bot.WebServer.PrivateRouter(), bot.WebServer.PublicRouter())
		}
	}
}
