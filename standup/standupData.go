package standup

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type standupData struct {
	Yesterday  string
	Today      string
	Blocking   string
	LastUpdate time.Time
}

func (sd standupData) String() string {
	str := fmt.Sprintf("Yesterday: %s\n", sd.Yesterday)
	str += fmt.Sprintf("Today: %s\n", sd.Today)
	str += fmt.Sprintf("Blocking: %s\n", sd.Blocking)
	return str
}

type standupMap map[standupDate]standupUsers

func (sm standupMap) Keys() standupDates {
	keys := make(standupDates, 0, len(sm))
	for k := range sm {
		keys = append(keys, k)
	}
	return keys
}

// Filter returns a copy of standupMap filtered by user fields [Name || Email]
func (sm standupMap) filterByEmail(email string) standupMap {
	fsm := make(standupMap)
	for date, users := range sm {
		fsm[date] = users.filterByEmail(email)
	}
	return fsm
}

func lineBreak(nchars int) string {
	line := ""
	for i := 0; i < nchars; i++ {
		line += "="
	}
	return line
}

/* String stringifies the map such that it will print:
Standup Report
 DATE
=====
username1
Yesterday: blah
Today: blah blah
Blocking: blah blah blah

username2
Yesterday: blah
Today: blah blah
Blocking: blah blah blah

DATE
====

...

Unless there is only 1 user in the map, in which case it does not repeat
the username and will look like

Standup Report for username
 DATE
=====
Yesterday: blah
Today: blah blah
Blocking: blah blah blah

DATE
====
Yesterday: blah
Today: blah blah
Blocking: blah blah blah

DATE
====
....

*/
func (sm standupMap) String() (str string) {
	sorted := sm.Keys()
	sort.Sort(sorted)

	// first pass detects if there are multiple users or a single user (email is used as unique ID)
	seenUsers := make(map[string]standupUser)
	var lastUser standupUser
	singleUserReport := false

	for _, sdate := range sorted {
		users := sm[sdate]
		for _, user := range users {
			seenUsers[user.Profile.Email] = user
			lastUser = user
		}
	}

	// write header depending on single or multiple user case
	if len(seenUsers) == 1 {
		singleUserReport = true
		str += fmt.Sprintf("Standup Report for %s\n", lastUser.Name)
	} else {
		str += "Standup Report\n"
	}

	// second pass stringifies the body and only prints user name if multiple users exist
	for _, sdate := range sorted {
		users := sm[sdate]
		str += fmt.Sprintf("%s\n", sdate.String())
		str += fmt.Sprintf("%s\n", lineBreak(len(sdate.String())))
		for _, user := range users {
			// if only single user don't repeatedly write name
			if !singleUserReport {
				str += fmt.Sprintf("%s\n", user.Name)
			}
			str += fmt.Sprintf("%s\n", user.data.String())
		}
	}

	// replace multiple newlines with a single newline at the end.
	str = strings.TrimRight(str, "\n") + "\n"
	return
}
