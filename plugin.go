package ahipbot

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
