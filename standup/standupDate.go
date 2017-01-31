package standup

import (
	"strconv"
	"time"
)

type standupDate struct {
	year  int
	month time.Month
	day   int
}

func getStandupDate(daysFromToday int) standupDate {
	d := time.Now().Add(time.Duration(daysFromToday) * 24 * time.Hour)
	return standupDate{
		year:  d.Year(),
		month: d.Month(),
		day:   d.Day(),
	}
}

func unixToStandupDate(unix int64) standupDate {
	d := time.Unix(unix, 0).UTC()
	return standupDate{
		year:  d.Year(),
		month: d.Month(),
		day:   d.Day(),
	}
}

func (sd standupDate) next() standupDate {
	current := time.Date(sd.year, sd.month, sd.day, 0, 0, 0, 0, time.Local)
	next := current.Add(24 * time.Hour)
	return standupDate{
		year:  next.Year(),
		month: next.Month(),
		day:   next.Day(),
	}
}

func (sd standupDate) String() string {
	return strconv.Itoa(sd.year) + "-" + sd.month.String() + "-" + strconv.Itoa(sd.day)
}

func (sd standupDate) Unix() int64 {
	return time.Date(sd.year, sd.month, sd.day, 0, 0, 0, 0, time.Local).Unix()
}

func (sd standupDate) toUnixUTCString() string {
	return strconv.FormatInt(sd.Unix(), 10)
}

type standupDates []standupDate

func (slice standupDates) Len() int {
	return len(slice)
}

func (slice standupDates) Less(i, j int) bool {
	return slice[i].Unix() < slice[j].Unix()
}

func (slice standupDates) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
