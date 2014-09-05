package plotbot

import "time"

//
// Bot plugins
//
type PluginConfig struct {
	// Whether to handle the bot's own messages
	EchoMessages bool

	// Whether to handle messages that are not destined to me
	OnlyMentions bool
}

type Plugin interface {
	// Handle handles incoming messages matching the constraints
	// from PluginConfig.
	Handle(*Bot, *BotMessage)
	Config() *PluginConfig
}

var registeredPlugins = make([]func(*Bot) Plugin, 0)

// RegisterPlugin defers loading of plugins until main() is called with
// config file, storage and environment ready.
func RegisterPlugin(newFunc func(*Bot) Plugin) {
	registeredPlugins = append(registeredPlugins, newFunc)
}

var loadedPlugins = make([]Plugin, 0)

func LoadPlugins(bot *Bot) {
	for _, newFunc := range registeredPlugins {
		loadedPlugins = append(loadedPlugins, newFunc(bot))
	}
}

//
// Web plugins
//
type WebHandler interface {
	Run()
}

var registeredWebHandler func(*Bot, []Plugin) WebHandler = nil
var loadedWebHandler WebHandler = nil

func RegisterWebHandler(newFunc func(*Bot, []Plugin) WebHandler) {
	registeredWebHandler = newFunc
}

func LoadWebHandler(bot *Bot) {
	if registeredWebHandler != nil {
		loadedWebHandler = registeredWebHandler(bot, loadedPlugins)
		go loadedWebHandler.Run()
	}
}

//
// Rewarder plugin
//
type Rewarder interface {
	RegisterBadge(shortName, title, description string)
	LogEvent(user *User, event string, data interface{}) error
	FetchLogsSince(user *User, since time.Time, event string, data interface{}) error
	FetchLastLog(user *User, event string, data interface{}) error
	FetchLastNLogs(user *User, num int, event string, data interface{}) error
	AwardBadge(bot *Bot, user *User, shortName string) error
}

var registeredRewarder func(*Bot) Rewarder = nil

func RegisterRewarder(newFunc func(*Bot) Rewarder) {
	registeredRewarder = newFunc
}

func LoadRewarder(bot *Bot) {
	if registeredRewarder != nil {
		bot.Rewarder = registeredRewarder(bot)
	}
}
