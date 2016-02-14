package recognition

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/abourget/slick"
	"github.com/nlopes/slack"
)

func (p *Plugin) listenRecognize() {
	p.bot.Listen(&slick.Listener{
		Matches:            regexp.MustCompile(`!recognize ((<@U[A-Z0-9]+(|[a-zA-Z0-9_-])?>(, ?| and )?)+) for (.*)`),
		MessageHandlerFunc: p.handleRecognize,
	})
}

func (p *Plugin) handleRecognize(listen *slick.Listener, msg *slick.Message) {
	users := msg.Match[1]
	feat := msg.Match[5]

	channel := p.bot.GetChannelByName(p.config.Channel)
	if channel == nil {
		fmt.Println("Didn't find the recognitions, can't handle `!recognition` requests. Searched for:", p.config.Channel)
		return
	}

	recipients := parseRecipients(users)

	if userIsInRecipients(msg.FromUser.ID, recipients) {
		msg.ReplyMention("you can't recognize yourself, can you ?!")
		return
	}

	announcement := p.bot.SendOutgoingMessage(fmt.Sprintf("<@%s|%s> would like to recognize %s\n>>> For %s", msg.FromUser.ID, msg.FromUser.Name, users, feat), channel.ID)

	announcement.AddReaction("+1")
	announcement.AddReaction("dart")
	announcement.AddReaction("tada")
	announcement.AddReaction("100")
	announcement.AddReaction("clap")
	announcement.AddReaction("muscle")
	announcement.AddReaction("nerd_face")
	announcement.AddReaction("joy")

	announcement.OnAck(func(ack *slack.AckMessage) {
		ts := ack.Timestamp
		domain := p.bot.Config.TeamDomain
		url := fmt.Sprintf("https://%s.slack.com/archives/%s/p%s", domain, channel.Name, strings.Replace(ts, ".", "", 1))
		msg.ReplyMention("Great! Everyone can upvote this recognition here %s", url)

		recog := &Recognition{
			MsgTimestamp: ts,
			CreatedAt:    time.Now(),
			Sender:       msg.FromUser.ID,
			Recipients:   recipients,
			Categories:   []string{},
			Reactions: map[string]int{
				msg.FromUser.ID: 1,
			},
		}
		p.store.Put(recog)

		p.bot.PubSub.Pub(recog, "recognition:recognized")

		//fmt.Println("Timestamp for the message:", ts)
	})
}

// parseRecipients transforms a string like "<@U123123|superbob>,
// <@U15125|mama> and <@U999|rita>" into []string{"U123123", "U15125",
// "U999"}
var recipientsRE = regexp.MustCompile(`@(U[A-Z0-9a-z]+)`)

func parseRecipients(input string) []string {
	results := recipientsRE.FindAllStringSubmatch(input, -1)
	if results == nil {
		log.Println("ERROR parsing recipients for recognition, couldn't find any user in the processed list:", input)
		return []string{}
	}

	var out []string
	for _, name := range results {
		out = append(out, name[1])
	}
	return out
}

func userIsInRecipients(user string, recipients []string) bool {
	for _, val := range recipients {
		if val == user {
			return true
		}
	}
	return false
}
