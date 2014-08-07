package ahipbot

import (
	"testing"

	"github.com/abourget/ahipbot/hipchatv2"
)

func TestBotGetUserByName(t *testing.T) {
	u1 := hipchatv2.User{Name: "Bob Dylan"}
	u2 := hipchatv2.User{Name: "Robert"}
	bot := &Bot{
		Users: []hipchatv2.User{u1, u2},
	}
	user := bot.GetUser("Bob Dylan")
	if user.Name != "Bob Dylan" {
		t.Error("Bob Dylan wasn't in there")
	}
}

func TestBotGetUserByGroupChatJID(t *testing.T) {
	u1 := hipchatv2.User{Name: "Alexandre Bourget", ID: 10}
	bot := &Bot{
		Users: []hipchatv2.User{u1},
	}
	user := bot.GetUser("123823_devops@conf.hipchat.com/Alexandre Bourget")
	if user == nil {
		t.Error("Didn't find user at all")
		return
	}
	if user.ID != 10 {
		t.Error("Couldn't get 'Alexandre Bourget'")
	}
}

func TestBotGetUserByPrivateChatJID(t *testing.T) {
	u1 := hipchatv2.User{JID: "123823_902463@chat.hipchat.com", ID: 10}
	bot := &Bot{
		Users: []hipchatv2.User{u1},
	}
	user := bot.GetUser("123823_902463@chat.hipchat.com/linux")
	if user == nil {
		t.Error("Didn't find user at all")
		return
	}
	if user.ID != 10 {
		t.Error("Couldn't get 'Alexandre Bourget'")
	}
}
