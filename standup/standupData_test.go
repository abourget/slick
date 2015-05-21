package standup

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/nlopes/slack"
)

func TestStandupDataString(t *testing.T) {

	datastr := standupData{
		Yesterday: "a",
		Today:     "b",
		Blocking:  "c",
	}.String()

	str := `Yesterday: a
Today: b
Blocking: c
`

	if datastr != str {
		t.Error("expected '" + datastr + "'" + " to be '" + str + "'")
	}
}

func getTestStandupMap() standupMap {

	sm := make(standupMap)

	uA := standupUser{
		&slack.User{
			Name:    "A",
			Profile: slack.UserProfile{Email: "A@test.ly"},
		},
		standupData{},
	}

	uB := standupUser{
		&slack.User{
			Name:    "B",
			Profile: slack.UserProfile{Email: "B@test.ly"},
		},
		standupData{},
	}

	var unixDate int64 = 1431921600 // 2015-May-18
	sds := []standupDate{unixToStandupDate(unixDate), unixToStandupDate(unixDate).next()}

	for i := 0; i < 2; i += 1 {
		uA.data.Yesterday = strconv.Itoa(i)
		uA.data.Today = strconv.Itoa(i)
		uA.data.Blocking = strconv.Itoa(i)
		uB.data.Yesterday = strconv.Itoa(i)
		uB.data.Today = strconv.Itoa(i)
		uB.data.Blocking = strconv.Itoa(i)

		sm[sds[i]] = standupUsers{uA, uB}
	}

	return sm
}

func TestSingleUserMapString(t *testing.T) {

	sm := getTestStandupMap().filterByEmail("B@test.ly")

	singleReport := fmt.Sprintf("%s", sm)
	expectedReport := `Standup Report for B
2015-May-18
===========
Yesterday: 0
Today: 0
Blocking: 0

2015-May-19
===========
Yesterday: 1
Today: 1
Blocking: 1
`

	if singleReport != expectedReport {
		t.Error("Expected '" + singleReport + "' to be '" + expectedReport + "'")
	}

}

func TestMultipleUserMapString(t *testing.T) {

	sm := getTestStandupMap()

	multiReport := fmt.Sprintf("%s", sm)
	expectedReport := `Standup Report
2015-May-18
===========
A
Yesterday: 0
Today: 0
Blocking: 0

B
Yesterday: 0
Today: 0
Blocking: 0

2015-May-19
===========
A
Yesterday: 1
Today: 1
Blocking: 1

B
Yesterday: 1
Today: 1
Blocking: 1
`

	if multiReport != expectedReport {
		t.Error("Expected '" + multiReport + "' to be '" + expectedReport + "'")
	}

}
