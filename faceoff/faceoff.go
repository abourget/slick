package faceoff

// Package faceoff implements a game to learn the faces and names of your team mates.
//
// Start with !faceoff in any channel and let the fun begin.

import (
	_ "image/jpeg"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/abourget/slick"
	"github.com/nlopes/slack"
)

func init() {
	slick.RegisterPlugin(&Faceoff{})
}

type Faceoff struct {
	sync.Mutex

	bot   *slick.Bot
	users map[string]*User
}

const faceoffKey = "/faceoff/users/stats"

func (p *Faceoff) InitPlugin(bot *slick.Bot) {
	p.bot = bot

	faceoffRE := regexp.MustCompile("^!face[ _-]?off")
	bot.Listen(&slick.Listener{
		PublicOnly:     true,
		Matches:        faceoffRE,
		ListenForEdits: true,
		MessageHandlerFunc: func(listen *slick.Listener, msg *slick.Message) {
			// Launch a new game

			g := &Game{
				Faceoff:         p,
				OriginalMessage: msg,
				Channel:         msg.FromChannel,
				Started:         time.Now(),
			}
			msg.Reply("Ok, are you ready ?")
			go func() {
				g.Launch()
			}()
		},
	})

	bot.Listen(&slick.Listener{
		PrivateOnly: true,
		Matches:     faceoffRE,
		MessageHandlerFunc: func(listen *slick.Listener, msg *slick.Message) {
			user := p.users[msg.FromUser.ID]
			if user == nil {
				return
			}
			msg.Reply("Your !faceoff scores:\n`" + user.ScoreLine() + "`")
		},
	})

	bot.Listen(&slick.Listener{
		PrivateOnly: true,
		Matches:     faceoffRE,
		EventHandlerFunc: func(listen *slick.Listener, ev interface{}) {
			if _, ok := ev.(*slack.HelloEvent); ok {
				log.Println("faceoff: loading data")
				_ = p.bot.GetDBKey(faceoffKey, &p.users)

				// on HELLO, once the bot has updated all its Users..
				p.updateUsersFromSlack()
				log.Printf("faceoff: got %d profiles\n", len(p.users))
				p.flushData()
			}
		},
	})
}

func (p *Faceoff) updateUsersFromSlack() {
	if p.users == nil {
		p.users = make(map[string]*User)
	}

	for _, slackUser := range p.bot.Users {
		if slackUser.IsBot || slackUser.Deleted || slackUser.IsUltraRestricted || slackUser.IsRestricted || slackUser.RealName == "slackbot" {
			delete(p.users, slackUser.ID)
			continue
		}

		_, found := p.users[slackUser.ID]
		if !found {
			p.users[slackUser.ID] = &User{
				ID: slackUser.ID,
			}
		}
	}
}

func (p *Faceoff) flushData() {
	err := p.bot.PutDBKey(faceoffKey, p.users)
	if err != nil {
		log.Println("Failed to flush Faceoff data!", err)
	}
}
