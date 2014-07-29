package main

import "github.com/tkawachi/hipchat"
import "strings"

type BotMessage struct {
	*hipchat.Message
	BotMentioned bool
}

func (msg *BotMessage) ContainsAnyCased(strs []string) bool {
	for _, s := range strs {
		if strings.Contains(msg.Body, s) {
			return true
		}
	}
	return false
}

func (msg *BotMessage) ContainsAny(strs []string) bool {
	lowerStr := strings.ToLower(msg.Body)

	for _, s := range strs {
		lowerInput := strings.ToLower(s)

		if strings.Contains(lowerStr, lowerInput) {
			return true
		}
	}
	return false
}

func (msg *BotMessage) Contains(s string) bool {
	lowerStr := strings.ToLower(msg.Body)
	lowerInput := strings.ToLower(s)

	if strings.Contains(lowerStr, lowerInput) {
		return true
	}
	return false
}

func (msg *BotMessage) Reply(s string) *BotReply {
	return &BotReply{
		To:      msg.From,
		Message: s,
	}
}
