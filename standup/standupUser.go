package standup

import (
	"strings"

	"github.com/nlopes/slack"
)

type standupUser struct {
	*slack.User
	data standupData
}

func (u standupUser) FirstName() string {
	return strings.Split(u.Name, " ")[0]
}

type standupUsers []standupUser

// return a slice copy of users filtered by users Email
// TODO make this less specific to email, pass in a user struct filter and match against fields
func (users standupUsers) filterByEmail(email string) (fusers standupUsers) {
	for _, user := range users {
		if email == user.Profile.Email {
			fusers = append(fusers, user)
		}
	}
	return
}
