package bugger

import (
	"fmt"
	"log"
	"time"

	"github.com/abourget/slick"
	"github.com/abourget/slick/github"
	"github.com/abourget/slick/util"
)

const dfltReportLength = 7 // days

func init() {
	slick.RegisterPlugin(&Bugger{})
}

type Bugger struct {
	bot      *slick.Bot
	ghclient github.Client
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

func (bugger *Bugger) InitPlugin(bot *slick.Bot) {

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

	bot.Listen(&slick.Listener{
		MessageHandlerFunc: bugger.ChatHandler,
	})

}

func (bugger *Bugger) ChatHandler(listen *slick.Listener, msg *slick.Message) {

	if !msg.MentionsMe {
		return
	}

	if msg.ContainsAny([]string{"bug report", "bug count"}) && msg.ContainsAny([]string{"how", "help"}) {

		var report string

		if msg.Contains("bug report") {
			report = "bug report"
		} else {
			report = "bug count"
		}
		mention := bugger.bot.Config.Nickname

		msg.Reply(fmt.Sprintf(
			`Usage: %s, [give me a | insert demand]  <%s>  [from the | syntax filler] [last | past] [n] [days | weeks]
examples: %s, please give me a %s over the last 5 days
%s, produce a %s   (7 day default)
%s, I want a %s from the past 2 weeks
%s, %s from the past week`, mention, report, mention, report, mention, report, mention, report, mention, report))

	} else if msg.Contains("bug report") {

		days := util.GetDaysFromQuery(msg.Text)
		bugger.messageReport(days, msg, listen, func() string {
			reporter := bugger.makeBugReporter(days)
			return reporter.printReport(days)
		})

	} else if msg.Contains("bug count") {

		days := util.GetDaysFromQuery(msg.Text)
		bugger.messageReport(days, msg, listen, func() string {
			reporter := bugger.makeBugReporter(days)
			return reporter.printCount(days)
		})

	}

	return

}

func (bugger *Bugger) messageReport(days int, msg *slick.Message, listen *slick.Listener, genReport func() string) {

	if days > 31 {
		msg.Reply(fmt.Sprintf("Whaoz, %d is too much data to compile - well maybe not, I am just scared", days))
		return
	}

	msg.Reply(bugger.bot.WithMood("Building report - one moment please",
		"Whaooo! Pinging those githubbers - Let's do this!"))

	msg.Reply(genReport())

}
