package main

import (
)


type Funny struct {
	config *PluginConfig
}

func NewFunny(bot *Hipbot) *Funny {
	return &Funny{config: &PluginConfig{
		EchoMessages: false,
		OnlyMentions: false,
	}}
}

// Configuration
func (funny *Funny) Config() *PluginConfig {
	return funny.config
}

// Handler
func (funny *Funny) Handle(bot *Hipbot, msg *BotMessage) {
	// Anywhere
	if msg.ContainsAny([]string{"what is your problem", "what's your problem"}) {
		bot.Reply(msg, "http://media4.giphy.com/media/19hU0m3TJe6I/200w.gif")
		return
	}

	// Only if we were mentioned
	if msg.BotMentioned {
		if msg.ContainsAny([]string{"excitement", "exciting"}) {
			bot.Reply(msg, "http://static.fjcdn.com/gifs/Japanese+kids+spongebob+toys_0ad21b_3186721.gif")
		}
	}
}
