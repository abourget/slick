package standup

import (
	"testing"

	"github.com/nlopes/slack"
)

func TestFilterByEmail(t *testing.T) {

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

	users := standupUsers{uA, uB}
	filteredUsers := users.filterByEmail("B@test.ly")

	if len(filteredUsers) != 1 {
		t.Error("expected filteredUsers to be length 1, instead got", len(filteredUsers))
	}

	if filteredUsers[0].Name != "B" {
		t.Error("expected filtered User Name to be 'B', instead got", filteredUsers[0].Name)
	}
}
