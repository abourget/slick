package main

type BotReply struct {
	To      string
	Message string
}

type PluginConfig struct {
	EchoMessages bool  // Whether to handle the bot's own messages
	OnlyMentions bool  // Whether to handle messages that are not destined to me
}

type Plugin interface {
	Handle(*Hipbot, *BotMessage)
	Config() *PluginConfig
}
