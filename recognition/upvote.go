package recognition

import (
	"log"
	"strings"

	"github.com/abourget/slick"
)

func (p *Plugin) listenUpvotes() {
	p.bot.Listen(&slick.Listener{
		EventHandlerFunc: func(_ *slick.Listener, event interface{}) {
			react := slick.ParseReactionEvent(event)
			if react == nil {
				return
			}

			log.Println("Fetching item ts:", react.Item.Timestamp)
			recognition := p.store.Get(react.Item.Timestamp)
			if recognition == nil {
				return
			}

			user := p.bot.Users[react.User]
			if user.IsBot {
				log.Println("Not taking votes from bots")
				return
			}

			if p.config.DomainRestriction != "" && !strings.HasSuffix(user.Profile.Email, p.config.DomainRestriction) {
				log.Printf("Not taking votes from people outsite domain %q, was %q", p.config.DomainRestriction, user.Profile.Email)
				return
			}

			log.Println("Up/down voting recognition")
			p.upvoteRecognition(recognition, react)
		},
	})
}

func (p *Plugin) upvoteRecognition(recognition *Recognition, reaction *slick.ReactionEvent) {
	direction := 1
	if reaction.Type == slick.ReactionRemoved {
		direction = -1
	}
	recognition.Reactions[reaction.User] += direction
	p.store.Put(recognition)
}
