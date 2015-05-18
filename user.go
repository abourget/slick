package slick

import "github.com/abourget/slick/hipchatv2"

type User struct {
	ID          int64  `json:"id"`
	JID         string `json:"xmpp_jid"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Title       string `json:"title"`
	PhotoURL    string `json:"photo_url"`
	MentionName string `json:"mention_name"`
}

func UserFromHipchatv2(user hipchatv2.User) (newUser User) {
	newUser.ID = user.ID
	newUser.JID = user.JID
	newUser.Name = user.Name
	newUser.Email = user.Email
	newUser.Title = user.Title
	newUser.PhotoURL = user.PhotoURL
	newUser.MentionName = user.MentionName
	return
}
