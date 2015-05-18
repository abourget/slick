package standup

import (
	"strings"

	"github.com/plotly/plotbot"
)

type standupUser struct {
	*plotbot.User
	data standupData
}

func (u standupUser) FirstName() string {
	return strings.Split(u.Name, " ")[0]
}
