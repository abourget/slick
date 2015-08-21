package standup

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/abourget/slick"
)

var sectionRegexp = regexp.MustCompile(`(?mi)^!(yesterday|today|blocking)`)

type sectionMatch struct {
	name string
	text string
}

// extractSectionAndText returns the "today", "this is what I did today" section from the result of "FindAllStringSubmatchindex" call.
func extractSectionAndText(input string, res [][]int) []sectionMatch {
	out := make([]sectionMatch, 0, 3)

	for i := 0; i < len(res); i++ {
		el := res[i]

		section := input[el[2]:el[3]] // (2,3) is second group's (start,end)
		strings.ToLower(section)

		var endFullText = len(input)
		if (i + 1) < len(res) {
			endFullText = res[i+1][0]
		}
		fullText := input[el[1]:endFullText]
		fullText = strings.TrimSpace(fullText)

		out = append(out, sectionMatch{section, fullText})
	}

	return out
}

func (standup *Standup) TriggerReminders(msg *slick.Message, section string) {
	standup.sectionUpdates <- sectionUpdate{section, msg}
}

//
// Reminder to complete all sections and reception confirmation message
//

func (standup *Standup) manageUpdatesInteraction() {
	remindCh := make(chan *slick.Message)
	resetCh := make(chan *slick.Message)

	for {
		select {
		case update := <-standup.sectionUpdates:
			userEmail := update.msg.FromUser.Profile.Email
			progress := userProgressMap[userEmail]
			if progress == nil {
				progress = &userProgress{
					sectionsDone: make(map[string]bool),
					cancelTimer:  make(chan bool),
				}
				userProgressMap[userEmail] = progress
				progress.sectionsDone[update.section] = true
				go progress.waitAndCheckProgress(update.msg, remindCh)
				go progress.waitForReset(update.msg, resetCh)
			} else {
				close(progress.cancelTimer)

				progress.sectionsDone[update.section] = true
				numDone := len(progress.sectionsDone)
				if numDone == 3 {
					update.msg.ReplyMention("got it!")
					delete(userProgressMap, update.msg.FromUser.Profile.Email)
				} else {
					progress.cancelTimer = make(chan bool)
					go progress.waitAndCheckProgress(update.msg, remindCh)
				}
			}

		case msg := <-resetCh:
			userEmail := msg.FromUser.Profile.Email
			progress := userProgressMap[userEmail]
			if progress != nil {
				close(progress.cancelTimer)
			}
			delete(userProgressMap, userEmail)

		case msg := <-remindCh:
			// Do the reminding for that user
			userEmail := msg.FromUser.Profile.Email
			userProgress := userProgressMap[userEmail]
			if userProgress == nil {
				continue
			}

			remains := make([]string, 0, 3)
			if userProgress.sectionsDone["today"] == false {
				remains = append(remains, "today")
			}
			if userProgress.sectionsDone["yesterday"] == false {
				remains = append(remains, "yesterday")
			}
			if userProgress.sectionsDone["blocking"] == false {
				remains = append(remains, "blocking stuff")
			}

			remain := strings.Join(remains, " or ")

			if remain != "" {
				msg.ReplyMention(fmt.Sprintf("what about %s ?", remain))
			}
		}
	}
}

type sectionUpdate struct {
	section string
	msg     *slick.Message
}

var userProgressMap = make(map[string]*userProgress)

type userProgress struct {
	sectionsDone map[string]bool
	cancelTimer  chan bool
}

func (up *userProgress) waitAndCheckProgress(msg *slick.Message, remindCh chan *slick.Message) {
	select {
	case <-time.After(90 * time.Second):
		remindCh <- msg
	case <-up.cancelTimer:
		return
	}
}

// waitForReset waits a couple of minutes and stops listening to that user altogether.  We want to poke the user once or twice if he's slow.. but not eternally.
func (up *userProgress) waitForReset(msg *slick.Message, resetCh chan *slick.Message) {
	<-time.After(15 * time.Minute)
	resetCh <- msg
}
