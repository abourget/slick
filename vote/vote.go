package vote

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/abourget/slick"
)

type Vote struct {
	bot          *slick.Bot
	runningVotes map[string][]vote // votes per channel
	mutex        sync.Mutex
}

func init() {
	slick.RegisterPlugin(&Vote{
		runningVotes: make(map[string][]vote),
	})
}

func (vote *Vote) InitPlugin(bot *slick.Bot) {
	vote.bot = bot

	bot.ListenFor(&slick.Conversation{
		PublicOnly:  true,
		HandlerFunc: vote.voteHandler,
	})
}

type vote struct {
	user string
	vote string
}

func (v *Vote) voteHandler(conv *slick.Conversation, msg *slick.Message) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	bot := v.bot
	// TODO:    ok, @kat wants to survey what's for lunch, use "!vote The Food Place http://food-place.url" .. you can vote for the same place with a substring: "!vote food place"
	// TODO: match "!vote Mucha Dogs http://bigdogs.com"
	// TODO: match "!vote mucha dogs"
	// TODO: match "!vote Other place

	if msg.Text == "!what-for-lunch" || msg.Text == "!vote-for-lunch" {
		bot.ReplyMention(msg, "you can say `!what-for-lunch 5m` to get a vote that will last 5 minutes. `!vote-for-lunch` is an alias")
		return
	}

	if msg.HasPrefix("!what-for-lunch ") || msg.HasPrefix("!vote-for-lunch ") {
		if v.runningVotes[msg.FromChannel.Id] != nil {
			bot.ReplyMention(msg, "vote is already running!")
			return
		}

		timing := strings.TrimSpace(strings.SplitN(msg.Text, " ", 2)[1])
		dur, err := time.ParseDuration(timing)
		if err != nil {
			bot.ReplyMention(msg, fmt.Sprintf("couldn't parse duration: %s", err))
			return
		}

		v.runningVotes[msg.FromChannel.Id] = make([]vote, 0)

		go func() {
			time.Sleep(dur)

			v.mutex.Lock()
			defer v.mutex.Unlock()

			res := make(map[string]int)
			for _, oneVote := range v.runningVotes[msg.FromChannel.Id] {
				res[oneVote.vote] = res[oneVote.vote] + 1
			}

			// TODO: print report, clear up
			if len(res) == 0 {
				bot.ReplyMention(msg, "polls closed, but no one voted")
			} else {
				out := []string{"polls closed, here are the results:"}
				for theVote, count := range res {
					plural := ""
					if count > 1 {
						plural = "s"
					}
					out = append(out, fmt.Sprintf("* %s: %d vote%s", theVote, count, plural))
				}
				bot.ReplyMention(msg, strings.Join(out, "\n"))
			}

			delete(v.runningVotes, msg.FromChannel.Id)
		}()

		bot.Reply(msg, "<!channel> okay, what do we eat ? Votes are open. Use `!vote The Food Place http://food-place.url` .. you can vote for the same place with a substring, ex: `!vote food place`")

	}

	if msg.HasPrefix("!vote ") {
		running := v.runningVotes[msg.FromChannel.Id]
		if running == nil {
			bot.Reply(msg, bot.WithMood("what vote ?!", "oh you're so cute! voting while there's no vote going on !"))
			return
		}

		voteCast := strings.TrimSpace(strings.SplitN(msg.Text, " ", 2)[1])
		if len(voteCast) == 0 {
			return
		}

		// TODO: check for dupe
		for _, prevVote := range running {
			if msg.FromUser.Id == prevVote.user {
				// buzz off if you voted already
				bot.ReplyMention(msg, bot.WithMood("you voted already", "trying to double vote ! how charming :)"))
				return
			}
		}

		for _, prevVote := range running {
			if strings.Contains(strings.ToLower(prevVote.vote), strings.ToLower(voteCast)) {
				running = append(running, vote{msg.FromUser.Id, prevVote.vote})
				v.runningVotes[msg.FromChannel.Id] = running
				bot.ReplyMention(msg, bot.WithMood("okay", "hmmm kaay"))
				return
			}
		}
		running = append(running, vote{msg.FromUser.Id, voteCast})
		v.runningVotes[msg.FromChannel.Id] = running
		bot.ReplyMention(msg, bot.WithMood("taking note", "taking note! what a creative mind..."))

		// TODO: match "!what-for-lunch 1h|5m|50s"

	}
}
