package slick

import (
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

func NextWeekdayTime(w time.Weekday, hour, min int) (time.Time, time.Duration) {
	t := time.Now().UTC()
	nowWeekday := t.Weekday()
	nowYear, nowMonth, nowDay := t.Date()

	delta := (int(w) - int(nowWeekday) + 7) % 7
	res := time.Date(nowYear, nowMonth, nowDay+delta, hour, min, 0, 0, t.Location())

	if res.Sub(t) <= 0 {
		res = res.AddDate(0, 0, 7)
	}
	return res, res.Sub(t)
}

func AfterNextWeekdayTime(w time.Weekday, hour, min int) <-chan time.Time {
	_, duration := NextWeekdayTime(w, hour, min)
	return time.After(duration)
}

func unixFromTimestamp(ts string) slack.JSONTime {
	i, _ := strconv.Atoi(strings.Split(ts, ".")[0])
	return slack.JSONTime(i)
}
