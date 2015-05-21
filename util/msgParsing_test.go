package util

import (
	"strconv"
	"testing"
)

type Query struct {
	text         string
	daysExpected int
	daysComputed int
}

func makeQuery(text string, daysExpected int) Query {
	return Query{
		text:         text,
		daysExpected: daysExpected,
		daysComputed: GetDaysFromQuery(text),
	}
}

func (q Query) notOk() bool {
	return q.daysExpected != q.daysComputed
}

func (q Query) toString() string {
	str := "expected '" + q.text + "' "
	str += "to yield " + strconv.Itoa(q.daysExpected) + " days "
	str += "and got " + strconv.Itoa(q.daysComputed)
	return str
}

func TestGetDaysFromQuery(t *testing.T) {

	if query := makeQuery("plot, give me a report for the past 5 days", 5); query.notOk() {
		t.Error(query.toString())
	}

	if query := makeQuery("plot, give me a report for the last 5 days", 5); query.notOk() {
		t.Error(query.toString())
	}

	if query := makeQuery("plot, give me a report for the past day", 1); query.notOk() {
		t.Error(query.toString())
	}

	if query := makeQuery("plot, give me a report for the last week", 7); query.notOk() {
		t.Error(query.toString())
	}

	if query := makeQuery("plot, give me a report for the past 2 week", 14); query.notOk() {
		t.Error(query.toString())
	}

	if query := makeQuery("plot, give me a report for today", 0); query.notOk() {
		t.Error(query.toString())
	}

	if query := makeQuery("first as tragedy, second as farce", 0); query.notOk() {
		t.Error(query.toString())
	}

	if query := makeQuery("plot, give me a report for this week", 7); query.notOk() {
		t.Error(query.toString())
	}

}
