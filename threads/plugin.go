package threads

import (
	"github.com/abourget/slick"
)

type Plugin struct {
	bot *slick.Bot
}

func init() {
	slick.RegisterPlugin(&Plugin{})
}

func (p *Plugin) InitPlugin(bot *slick.Bot) {
	p.bot = bot

	p.listenThreads()
}
