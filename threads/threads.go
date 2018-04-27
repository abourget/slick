package threads

import (
	"github.com/abourget/slick"
)

func (p *Plugin) listenThreads() {
	p.bot.Listen(&slick.Listener{
		MessageHandlerFunc: p.handleThreads,
	})
}

func (p *Plugin) handleThreads(listen *slick.Listener, msg *slick.Message) {
	if msg.ThreadTimestamp != "" {
		msg.Reply("Can I haz no threadz plz!")
	}
}
