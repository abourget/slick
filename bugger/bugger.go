package bugger

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/plotly/plotbot"
	"github.com/plotly/plotbot/github"
)

func init() {
	plotbot.RegisterPlugin(&Bugger{})
}

type Bugger struct {
	bot      *plotbot.Bot
	ghclient github.Client
}

type bugReport struct {
	bugs []github.IssueItem
}

func (r *bugReport) addBug(issue github.IssueItem) {
	r.bugs = append(r.bugs, issue)
}

func (r *bugReport) printReport(days int) (report string) {

	dayheader := fmt.Sprintf(" BUG REPORT FOR LAST %d DAYS ", days) // 20 spaces
	bar := "************************"

	report = fmt.Sprintf("/quote " + bar + dayheader + bar + "\n")

	report += fmt.Sprintf("|%-45s|%-7s|%-18s|\n", "bug title", "number", "squasher")
	title := ""
	for _, bug := range r.bugs {
		if len(bug.Title) > 45 {
			title = bug.Title[0:42] + "..."
		} else {
			title = bug.Title
		}
		report += fmt.Sprintf("|%-45s|%-7d|%-18s|\n", title, bug.Number, bug.LastClosedBy())
	}

	return
}

func (bugger *Bugger) getBugReport(days int) string {

	repo := bugger.ghclient.Conf.Repos[0]

	query := github.SearchQuery{
		Repo:        repo,
		Labels:      []string{"bug"},
		ClosedSince: time.Now().Add(-time.Duration(days) * (24 * time.Hour)).Format("2006-01-02"),
	}

	issueList, err := bugger.ghclient.DoSearchQuery(query)
	if err != nil {
		log.Print(err)
		return ""
	}

	/*
	 * Get an array of issues matching Filters
	 */
	issueChan := make(chan github.IssueItem, 1)
	go bugger.ghclient.DoEventQuery(issueList, issueChan)

	reporter := bugReport{}

	for issue := range issueChan {
		reporter.addBug(issue)
	}

	return reporter.printReport(days)
}

func (bugger *Bugger) InitChatPlugin(bot *plotbot.Bot) {

	/*
	 * Get an array of issues matching Filters
	 */
	bugger.bot = bot

	var conf struct {
		Github github.Conf
	}

	bot.LoadConfig(&conf)

	bugger.ghclient = github.Client{
		Conf: conf.Github,
	}

	bot.ListenFor(&plotbot.Conversation{
		HandlerFunc: bugger.ChatHandler,
	})

}

func (bugger *Bugger) ChatHandler(conv *plotbot.Conversation, msg *plotbot.Message) {

	if msg.MentionsMe && msg.Contains("bug report") {

		re := regexp.MustCompile(".*(?:last|past) (\\d+)?\\s?(day|week).*")
		hits := re.FindStringSubmatch(msg.Body)

		var days, weeks int
		var err error

		if len(hits) == 3 {
			howmany := hits[1]
			dayOrWeek := hits[2]

			if dayOrWeek == "day" {
				if howmany == "" {
					days = 7
				} else {
					days, err = strconv.Atoi(howmany)
					if err != nil {
						days = 7
					}
				}
			} else {
				if howmany == "" {
					days = 7
				} else {
					weeks, err = strconv.Atoi(howmany)
					if err != nil {
						days = 7
					} else {
						days = 7 * weeks
					}
				}
			}
		} else {
			days = 7
		}

		if days > 31 {
			conv.Reply(msg, fmt.Sprintf("Whaoz, %d is too much data to compile - well maybe not, I am just scared", days))
			return
		}

		conv.Reply(msg, fmt.Sprintf("hang on - let me ping those github kids"))

		bugreport := bugger.getBugReport(days)

		conv.Reply(msg, bugreport)
	}
	return

}
