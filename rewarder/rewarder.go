package rewarder

import "github.com/plotly/plotbot"

/**
 * This package implements the Rewards, Archievements, Badges and Events
 * framework
 **/

type Rewarder struct {
	badges []*Badge
}

func init() {
	plotbot.RegisterRewarder(func(bot *plotbot.Bot) plotbot.Rewarder {
		rew := &Rewarder{
			badges: make([]*Badge, 3),
		}
		rew.RegisterBadge("small_mentioner", "Small Mentioner", `This badge is awarded when you mention Plotbot's name for the first time.`)
		rew.RegisterBadge("intimate_mentioner", "Intimave Mentioner", `This badge is awarded when you talk to Plotbot privately for the first time`)
		rew.RegisterBadge("great_mentioner", "Great Mentioner", `This badge is awarded when you have mentioned Plotbot's name at least 10 times in the past week`)
		return rew
	})
}

func (rew *Rewarder) RegisterBadge(shortName, title, description string) {
	badge := &Badge{
		ShortName:   shortName,
		Title:       title,
		Description: description,
	}
	rew.badges = append(rew.badges, badge)
}

func (rew *Rewarder) AwardBadge(bot *plotbot.Bot, user *plotbot.User, shortName string) error {
	return nil
}
