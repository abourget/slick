package bugger

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/plotly/plotbot"
	"github.com/plotly/plotbot/github"
	"github.com/plotly/plotbot/utils"
)

func init() {
	plotbot.RegisterPlugin(&Bugger{})
}

type Bugger struct {
	bot      *plotbot.Bot
	ghclient github.Client
}

type bugReporter struct {
	bugs    []github.IssueItem
	Git2Hip map[string]string
}

func (r *bugReporter) addBug(issue github.IssueItem) {
	r.bugs = append(r.bugs, issue)
}

func (r *bugReporter) printReport(days int) (report string) {

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

func (r *bugReporter) printCount(days int) (count string) {

	dayheader := fmt.Sprintf(" BUG COUNT FOR LAST %d DAYS ", days) // 20 spaces
	bar := "*************"

	count = fmt.Sprintf("/quote " + bar + dayheader + bar + "\n")
	count += fmt.Sprintf("|%-30s|%-20s|\n", "team member", "number squashed")

	bugcount := make(map[string]int)

	for _, bug := range r.bugs {
		bugcount[bug.LastClosedBy()]++
	}

	for _, ghname := range util.SortedKeys(bugcount) {
		count += fmt.Sprintf("|%-30s|%-20d|\n", ghname, bugcount[ghname])
	}

	return

}

func (bugger *Bugger) makeBugReporter(days int) (reporter bugReporter) {

	repo := bugger.ghclient.Conf.Repos[0]

	query := github.SearchQuery{
		Repo:        repo,
		Labels:      []string{"bug"},
		ClosedSince: time.Now().Add(-time.Duration(days) * (24 * time.Hour)).Format("2006-01-02"),
	}

	issueList, err := bugger.ghclient.DoSearchQuery(query)
	if err != nil {
		log.Print(err)
		return
	}

	/*
	 * Get an array of issues matching Filters
	 */
	issueChan := make(chan github.IssueItem, 1)
	go bugger.ghclient.DoEventQuery(issueList, repo, issueChan)

	reporter.Git2Hip = bugger.ghclient.Conf.Github2Hipchat

	for issue := range issueChan {
		reporter.addBug(issue)
	}

	return
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

		days := getDaysFromQuery(msg.Body)

		if days > 31 {
			conv.Reply(msg, fmt.Sprintf("Whaoz, %d is too much data to compile - well maybe not, I am just scared", days))
			return
		}

		conv.Reply(msg, fmt.Sprintf("hang on - let me ping those github kids"))

		reporter := bugger.makeBugReporter(days)
		conv.Reply(msg, reporter.printReport(days))

	} else if msg.MentionsMe && msg.Contains("bug count") {

		days := getDaysFromQuery(msg.Body)

		if days > 31 {
			conv.Reply(msg, fmt.Sprintf("Whaoz, %d is too much data to compile - well maybe not, I am just scared", days))
			return
		}

		conv.Reply(msg, fmt.Sprintf("hang on - let me ping those github kids"))

		reporter := bugger.makeBugReporter(days)

		conv.Reply(msg, reporter.printCount(days))
	}

	return

}

func getDaysFromQuery(text string) (days int) {

	re := regexp.MustCompile(".*(?:last|past) (\\d+)?\\s?(day|week).*")
	hits := re.FindStringSubmatch(text)

	var weeks int
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

	return
}
