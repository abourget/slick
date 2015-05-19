package bugger

import (
	"fmt"

	"github.com/plotly/plotbot/github"
	"github.com/plotly/plotbot/util"
)

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
	bar := "***"

	count = fmt.Sprintf("/quote " + bar + dayheader + bar + "\n")
	count += fmt.Sprintf("|%-20s|%-10s|\n", "team member", "# squashed")

	bugcount := make(map[string]int)

	for _, bug := range r.bugs {
		bugcount[bug.LastClosedBy()]++
	}

	for _, ghname := range util.SortedKeys(bugcount) {
		count += fmt.Sprintf("|%-20s|%-10d|\n", ghname, bugcount[ghname])
	}

	total := 0
	for _, value := range bugcount {
		total += value
	}

	count += fmt.Sprintf("|%-20s|%-10d|\n", "TOTAL", total)

	return

}
