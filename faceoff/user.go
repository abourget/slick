package faceoff

import "fmt"

func (p *Faceoff) UpdateUsersWithChallengeResults(c *Challenge) {
	for user, idx := range c.Replies {
		u := p.users[user]
		if u == nil {
			continue
		}

		if idx == c.RightAnswerIndex {
			u.RightAnswers++
		} else {
			u.WrongAnswers++
		}
	}

	u := p.users[c.FirstCorrectReply]
	if u != nil && len(c.Replies) > 1 {
		u.Fastest++
	}

	for idx, shown := range c.UsersShown {
		u := p.users[shown]
		if u == nil {
			continue
		}

		u.Shown++

		if c.RightAnswerIndex == idx {
			u.LookedFor++
			if c.FirstCorrectReply != "" {
				u.Found++
			}
		}
	}

	p.ComputeUserScores()
}

type User struct {
	// `slick.GetUser` ID
	ID string
	// Count of good answers, fastest or not.
	RightAnswers int
	// Count of bad answers,
	WrongAnswers int
	// Number of times the user was right and was the fastest.
	Fastest int

	// FoundUsers is a map to be used for completeness score, in next
	// version.
	FoundUsers map[string]bool

	// Number of times this user was looked for
	LookedFor int
	// Number of times this user was shown but wasn't the person we looked for.
	Shown int
	// Number of times this user was found
	Found int

	// PerformanceScore is a score composed of accuracy and speed
	// (successRate and fastest), relatively to others.
	PerformanceScore int
	// RankPosition is the position of this user relatively to other
	// users.  It is possible that several users are at the same
	// ranking position, if they have the same score.
	RankPosition int
	// RankAgainst is the total number of players
	RankAgainst int

	// Whether this users is excluded from the game.
	Excluded bool
}

// SuccessRate returns a percentage of the time the user was right, as
// compared to his bad answers.
func (u *User) SuccessRate() int {
	allAnswers := float64(u.RightAnswers) + float64(u.WrongAnswers)
	if allAnswers == 0 {
		return 0
	}
	return int(100 * float64(u.RightAnswers) / allAnswers)
}

func (user *User) ScoreLine() (out string) {
	// Fatest: 10 times, Right answer: 50%, Rank: 1 / 128
	out = fmt.Sprintf("Fastest: %d times -- Right answer: %d%%", user.Fastest, user.SuccessRate())

	if user.RankAgainst == 0 {
		return
	}

	// show rank only for the top 20% of players
	if (float64(user.RankPosition) / float64(user.RankAgainst)) < 0.2 {
		out = fmt.Sprintf("%s -- Rank: %d / %d", out, user.RankPosition, user.RankAgainst)
	}

	return
}
