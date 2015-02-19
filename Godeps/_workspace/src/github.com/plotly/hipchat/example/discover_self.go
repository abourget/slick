package main

import (
	"fmt"
	"github.com/daneharrigan/hipchat"
)

func main() {
	user := "11111_22222"
	pass := "secret"
	resource := "bot"

	client, err := hipchat.NewClient(user, pass, resource)
	if err != nil {
		fmt.Printf("client error: %s\n", err)
		return
	}

	var fullName string
	var mentionName string

	for _, user := range client.Users() {
		if user.Id == client.Id {
			fullName = user.Name
			mentionName = user.MentionName
			break
		}
	}

	fmt.Printf("name: %s\n", fullName)
	fmt.Printf("mention: %s\n", mentionName)
}
