package faceoff

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/abourget/slick"
	"github.com/nlopes/slack"
)

//
// Leaderboard plz! (mabeauchamp, etienne)
// Fastest timing !
// Numbers on a previous line ??
// Image after.. so that everyone sees the reactions before..

// Game is a run of multiple challenges in channel
type Game struct {
	// Channel in which the game is running
	Faceoff         *Faceoff
	OriginalMessage *slick.Message
	Channel         *slick.Channel
	Started         time.Time
	GameCount       int // number of games we did
	Challenges      []*Challenge
}

func (g *Game) Launch() {
	g.GameCount++
	if g.GameCount > 100 {
		log.Println("Ran enough Faceoff games.. stopping this run until we're asked again")
		return
	}

	c := newChallenge()
	g.Challenges = append(g.Challenges, c)

	c.PickUsers(g.Faceoff.users)

	var profileURLs []string
	var lookedForUser slack.User
	for idx, userID := range c.UsersShown {
		u := g.Faceoff.bot.Users[userID]
		if u.ID == "" {
			log.Println("faceoff: error finding user with ID", userID)
			g.OriginalMessage.Reply("error finding user with ID %q", userID)
			return
		}

		profileURLs = append(profileURLs, u.Profile.Image192)
		if idx == c.RightAnswerIndex {
			lookedForUser = u
		}
	}

	pngContent, err := c.BuildImage(profileURLs)
	if err != nil {
		log.Println("error building faceoff image:", err)
		go func() {
			time.Sleep(100 * time.Millisecond)
			g.Launch()
		}()
		return
	}

	// Trigger a line with who we're looking for..
	// Add the reactions, slowly, in order..
	// Send the image, after a good second..
	prepared := g.OriginalMessage.Reply(fmt.Sprintf("---\nBe prepared! We're looking for *%s* in the next image:", lookedForUser.RealName))
	prepared.OnAck(func(ev *slack.AckMessage) {
		go func() {
			delay := 750 * time.Millisecond
			g.Faceoff.bot.Slack.AddReaction("one", slack.NewRefToMessage(prepared.Channel, ev.Timestamp))
			time.Sleep(delay)
			g.Faceoff.bot.Slack.AddReaction("two", slack.NewRefToMessage(prepared.Channel, ev.Timestamp))
			time.Sleep(delay)
			g.Faceoff.bot.Slack.AddReaction("three", slack.NewRefToMessage(prepared.Channel, ev.Timestamp))
			time.Sleep(delay)
			g.Faceoff.bot.Slack.AddReaction("four", slack.NewRefToMessage(prepared.Channel, ev.Timestamp))

			g.showChallenge(c, lookedForUser, pngContent, ev.Timestamp)
		}()
	})

}

func (g *Game) showChallenge(c *Challenge, lookedForUser slack.User, pngContent []byte, ts string) {
	err := ioutil.WriteFile("/tmp/faceoff.png", pngContent, 0644)
	if err != nil {
		g.OriginalMessage.Reply("error writing temp faceoff image: %s", err)
		return
	}

	fmt.Println("*************************** before")

	_, err = g.Faceoff.bot.Slack.UploadFile(slack.FileUploadParameters{
		File:           "/tmp/faceoff.png",
		Filetype:       "png",
		Title:          fmt.Sprintf("Find: %s", lookedForUser.RealName),
		Channels:       []string{g.Channel.ID},
	})
	if err != nil {
		g.OriginalMessage.Reply("error uploading faceoff image: %s", err)
		return
	}

	// the FileCreatd giving us the timestamps arrive BEFORE this UploadFile call returns!
	// so we'd need to keep a few FileCreated thing.. for a few seconds, and then ...
	// link them together.

	g.Faceoff.bot.ListenReaction(ts, &slick.ReactionListener{
		ListenDuration: 60 * time.Second,
		Type:           slick.ReactionAdded,
		HandlerFunc: func(listen *slick.ReactionListener, ev *slick.ReactionEvent) {
			log.Println("*************************************** HANDLING USER REPLY")
			idx := -1
			switch ev.Emoji {
			case "one":
				idx = 0
			case "two":
				idx = 1
			case "three":
				idx = 2
			case "four":
				idx = 3
			}
			if idx == -1 {
				return
			}

			log.Println("Handling user reply", ev.Emoji, ev.User)

			c.HandleUserReply(ev.User, idx)

			listen.ResetNewDuration(10 * time.Second)
		},
		TimeoutFunc: func(listen *slick.ReactionListener) {
			defer listen.Close()

			if len(c.Replies) == 0 {
				g.OriginalMessage.Reply("oh well, I guess no one wanted to play!")
				// TODO: remove the original image
				return
			}

			g.Faceoff.UpdateUsersWithChallengeResults(c)

			g.Faceoff.flushData()

			if c.FirstCorrectReply != "" {
				user := g.Faceoff.users[c.FirstCorrectReply]
				// TODO:
				var whowaswho []string
				numbers := []string{":one:", ":two:", ":three:", ":four:"}
				for i := 0; i < 4; i++ {
					user := g.Faceoff.bot.GetUser(c.UsersShown[i])
					if user != nil {
						whowaswho = append(whowaswho, fmt.Sprintf("%s was *%s*", numbers[i], user.RealName))
					}
				}
				g.OriginalMessage.Reply("Congrats <@%s> ! You found <@%s> the fastest.\n%s\nYour scores: `%s`", c.FirstCorrectReply, c.UsersShown[c.RightAnswerIndex], strings.Join(whowaswho, ", "), user.ScoreLine())
			} else {
				g.OriginalMessage.Reply("No one found out !? Try again !")
			}

			go g.Launch()
		},
	})
}
