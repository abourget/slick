package rewarder

import "github.com/plotly/plotbot"

/**
 * This package implements the Rewards, Archievements, Badges and Events
 * framework
 **/

type Rewarder struct {
	badges []*Badge
	logs map[string][]*Event
}

func init() {
	plotbot.RegisterPlugin(&Rewarder{})
}

func (rew *Rewarder) InitRewarder(bot *plotbot.Bot) {
	rew.badges = make([]*Badge, 0)
	rew.logs = make(map[string][]*Event)
	rew.RegisterBadge("small_mentioner", "Small Mentioner", `This badge is awarded when you mention Plotbot's name for the first time.`)
	rew.RegisterBadge("intimate_mentioner", "Intimave Mentioner", `This badge is awarded when you talk to Plotbot privately for the first time`)
	rew.RegisterBadge("great_mentioner", "Great Mentioner", `This badge is awarded when you have mentioned Plotbot's name at least 10 times in the past week`)
}

func (rew *Rewarder) RegisterBadge(shortName, title, description string) {
	badge := &Badge{
		ShortName:   shortName,
		Title:       title,
		Description: description,
	}
	rew.badges = append(rew.badges, badge)
}

func (rew *Rewarder) Badges() []*Badge {
	return rew.badges
}

func (rew *Rewarder) AwardBadge(bot *plotbot.Bot, user *plotbot.User, shortName string) error {
	return nil
}
