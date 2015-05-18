package standup

import (
	"fmt"
	"sort"
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

type standupMap map[standupDate][]standupUser

func (sm standupMap) Keys() standupDates {
	keys := make(standupDates, 0, len(sm))
	for k := range sm {
		keys = append(keys, k)
	}
	return keys
}

func lineBreak(nchars int) string {
	line := ""
	for i := 0; i < nchars; i++ {
		line += "="
	}
	return line
}

func (sm standupMap) String() (str string) {
	sorted := sm.Keys()
	sort.Sort(sorted)

	for _, sdate := range sorted {
		users := sm[sdate]
		str += fmt.Sprintf("%s\n", sdate.String())
		str += fmt.Sprintf("%s\n", lineBreak(len(sdate.String())))
		for _, user := range users {
			str += fmt.Sprintf("%s\n", user.Name)
			str += fmt.Sprintf("%s\n", user.data.String())
		}
	}
	return
}
