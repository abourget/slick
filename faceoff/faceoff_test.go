package faceoff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoresComputing(t *testing.T) {
	f := &Faceoff{users: map[string]*User{
		"u1": {Fastest: 2, RightAnswers: 3, WrongAnswers: 1},
		"u2": {Fastest: 1, RightAnswers: 2, WrongAnswers: 2},
		"u3": {Fastest: 5, RightAnswers: 4},
	}}

	f.ComputeUserScores()

	assert.Equal(t, 2, f.users["u1"].Fastest)
	assert.Equal(t, 75, f.users["u1"].SuccessRate())
	assert.Equal(t, 3, f.users["u1"].RankAgainst)
	assert.Equal(t, 2, f.users["u1"].RankPosition)
}

func TestCalculatePerformance(t *testing.T) {
	assert.Equal(t, 1083, calculatePerformance(3, 20, 2, 15))
	assert.Equal(t, 1500, calculatePerformance(3, 20, 3, 20))
	assert.Equal(t, 666, calculatePerformance(6, 30, 2, 15))
	assert.Equal(t, 916, calculatePerformance(6, 30, 3, 20))
	assert.Equal(t, 0, calculatePerformance(3, 20, 0, 0))
	assert.Equal(t, 0, calculatePerformance(1, 1, 0, 0))
}

func TestScoresDisplay(t *testing.T) {
	user := &User{
		Fastest:      2,
		RightAnswers: 20,
		WrongAnswers: 80,
		RankPosition: 4,
		RankAgainst:  128,
	}

	assert.Equal(t, "Fastest: 2 times -- Right answer: 20% -- Rank: 4 / 128", user.ScoreLine())
}

func TestScoresDisplayWithoutRank(t *testing.T) {
	user := &User{
		Fastest:      2,
		RightAnswers: 20,
		WrongAnswers: 80,
		RankPosition: 127, // not so good, let's not say the rank :)
		RankAgainst:  128,
	}

	assert.Equal(t, "Fastest: 2 times -- Right answer: 20%", user.ScoreLine())
}

func TestUpdateUsersWithChallengeResults(t *testing.T) {
	f := &Faceoff{users: map[string]*User{
		"u1": {},
		"u2": {},
		"u3": {},
		"u4": {},
		"u5": {},
		"u6": {},
		"u7": {},
		"u8": {},
	}}
	c := &Challenge{
		UsersShown:       []string{"u1", "u2", "u3", "u4"},
		RightAnswerIndex: 1, // u2
		Replies: map[string]int{
			"u5": 1,
			"u6": 3,
			"u7": 0,
			"u8": 1,
		},
		FirstCorrectReply: "u5",
	}

	f.UpdateUsersWithChallengeResults(c)

	assert.Equal(t, 1, f.users["u1"].Shown)
	assert.Equal(t, 1, f.users["u2"].Shown)
	assert.Equal(t, 1, f.users["u3"].Shown)
	assert.Equal(t, 1, f.users["u4"].Shown)
	assert.Equal(t, 0, f.users["u5"].Shown)

	assert.Equal(t, 0, f.users["u1"].Found)
	assert.Equal(t, 1, f.users["u2"].Found)
	assert.Equal(t, 0, f.users["u5"].Found)

	assert.Equal(t, 0, f.users["u1"].RightAnswers)
	assert.Equal(t, 0, f.users["u1"].WrongAnswers)

	assert.Equal(t, 1, f.users["u5"].RightAnswers)
	assert.Equal(t, 0, f.users["u5"].WrongAnswers)
	assert.Equal(t, 0, f.users["u6"].RightAnswers)
	assert.Equal(t, 1, f.users["u6"].WrongAnswers)
	assert.Equal(t, 0, f.users["u7"].RightAnswers)
	assert.Equal(t, 1, f.users["u7"].WrongAnswers)
	assert.Equal(t, 1, f.users["u8"].RightAnswers)
	assert.Equal(t, 0, f.users["u8"].WrongAnswers)

	assert.Equal(t, 1, f.users["u5"].Fastest)
	assert.Equal(t, 0, f.users["u6"].Fastest)
}

func TestBuildImage(t *testing.T) {
	c := &Challenge{}

	_, err := c.BuildImage([]string{
		"https://avatars.slack-edge.com/2013-12-15/2156040809_192.jpg",
		"https://avatars.slack-edge.com/2016-03-14/26651220931_92fe2c6c1fe1735afa8c_192.jpg",
		"https://avatars.slack-edge.com/2016-03-11/26182082134_269648f19b40d8a87878_192.jpg",
		"https://avatars.slack-edge.com/2014-12-08/3167931031_42ef453717f47b15aa3b_192.jpg",
	})

	assert.NoError(t, err)
}

func TestChallengePickUsers(t *testing.T) {
	c := newChallenge()

	c.PickUsers(map[string]*User{
		"u1": {PerformanceScore: 20},
		"u2": {PerformanceScore: 40},
	})

	assert.Len(t, c.UsersShown, 2)
	assert.NotZero(t, c.RightAnswerIndex)
}
