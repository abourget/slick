package plotbot

import "testing"

func TestBotGetUserByName(t *testing.T) {
	u1 := User{Name: "Bob Dylan"}
	u2 := User{Name: "Robert"}
	bot := &Bot{
		Users: []User{u1, u2},
	}
	user := bot.GetUser("Bob Dylan")
	if user.Name != "Bob Dylan" {
		t.Error("Bob Dylan wasn't in there")
	}
}

func TestBotGetUserByGroupChatJID(t *testing.T) {
	u1 := User{Name: "Alexandre Bourget", ID: 10}
	bot := &Bot{
		Users: []User{u1},
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
	u1 := User{JID: "123823_902463@chat.hipchat.com", ID: 10}
	bot := &Bot{
		Users: []User{u1},
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

func TestBotGetRoomByName(t *testing.T) {
	r1 := Room{JID: "123823_devops@conf.hipchat.com", ID: 5, Name: "DevOps"}
	bot := &Bot{
		Rooms: []Room{r1},
	}

	room := bot.GetRoom("DevOps")

	if *room != r1 {
		t.Error("Didn't retrieve room")
	}
}

// PUBLIC: 2014/08/16 00:39:03 MESSAGE &hipchat.Message{From:"123823_devops@conf.hipchat.com/Alexandre Bourget", To:"123823_998606@chat.hipchat.com/bot", Body:"boo", MentionName:"", Type:"groupchat", Delay:(*hipchat.Delay)(nil)}
//    From: 123823_devops@conf.hipchat.com/Alexandre Bourget
// PRIVATE: 2014/08/16 00:39:06 MESSAGE &hipchat.Message{From:"123823_902438@chat.hipchat.com/linux", To:"123823_998606@chat.hipchat.com", Body:"moo", MentionName:"", Type:"chat", Delay:(*hipchat.Delay)(nil)}
//    From: 123823_902438@chat.hipchat.com/linux

func TestBotGetRoomByJID(t *testing.T) {
	r1 := Room{JID: "123823_devops@conf.hipchat.com", ID: 5, Name: "DevOps"}
	bot := &Bot{
		Rooms: []Room{r1},
	}

	room := bot.GetRoom("123823_devops@conf.hipchat.com/Alexandre Bourget")

	if *room != r1 {
		t.Error("Didn't retrieve room")
	}
}
