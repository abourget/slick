package slick

import (
	"time"

	"github.com/nlopes/slack"
)

// Channel is an abstraction of the Slack channels, with merged and
// distinguished data from IMs, Groups and Channels.
type Channel struct {
	ID       string
	Created  time.Time
	IsOpen   bool
	LastRead string
	Name     string
	Creator  string

	// Three mutually exclusives
	IsChannel bool
	IsGroup   bool
	IsIM      bool

	// Only with `IsChannel`
	IsGeneral bool
	Members   []string

	// Only for `IsGroup` (?)
	IsMember   bool
	IsArchived bool

	// Only for `IsIM`
	User          string
	IsUserDeleted bool

	// Only for `IsChannel || IsGroup`
	Topic   slack.Topic
	Purpose slack.Purpose
}

func ChannelFromSlackGroup(group slack.Group) Channel {
	return Channel{
		ID:       group.ID,
		Created:  group.Created.Time(),
		IsOpen:   group.IsOpen,
		LastRead: group.LastRead,
		Name:     group.Name,
		Creator:  group.Creator,
		Members:  group.Members,
		//IsMember:   group.IsMember,  wh00ps, not there anymore.
		IsArchived: group.IsArchived,
		Topic:      group.Topic,
		Purpose:    group.Purpose,
		IsGroup:    true,
	}
}

func ChannelFromSlackChannel(channel slack.Channel) Channel {
	return Channel{
		ID:         channel.ID,
		Created:    channel.Created.Time(),
		IsOpen:     channel.IsOpen,
		LastRead:   channel.LastRead,
		Name:       channel.Name,
		Creator:    channel.Creator,
		Members:    channel.Members,
		IsMember:   channel.IsMember,
		IsArchived: channel.IsArchived,
		Topic:      channel.Topic,
		Purpose:    channel.Purpose,
		IsChannel:  true,
	}
}
func ChannelFromSlackIM(im slack.IM) Channel {
	return Channel{
		ID:            im.ID,
		Created:       im.Created.Time(),
		IsOpen:        im.IsOpen,
		LastRead:      im.LastRead,
		Name:          im.User,
		User:          im.User,
		IsUserDeleted: im.IsUserDeleted,
		IsIM:          true,
	}
}
