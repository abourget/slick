package faceoff

import "sort"

func (p *Faceoff) ComputeUserScores() {
	fastestOfAll := 1
	highestSuccessRate := 1
	for _, u := range p.users {
		if u.Fastest > fastestOfAll {
			fastestOfAll = u.Fastest
		}
		rate := u.SuccessRate()
		if rate > highestSuccessRate {
			highestSuccessRate = rate
		}
	}

	var perfs []int
	for _, u := range p.users {
		perfs = append(perfs, calculatePerformance(fastestOfAll, highestSuccessRate, u.Fastest, u.SuccessRate()))
	}

	sort.Sort(sort.Reverse(sort.IntSlice(perfs)))

	for _, u := range p.users {
		u.PerformanceScore = calculatePerformance(fastestOfAll, highestSuccessRate, u.Fastest, u.SuccessRate())

		// Find its rank
		for idx, val := range perfs {
			if u.PerformanceScore >= val {
				u.RankAgainst = len(perfs)
				u.RankPosition = idx + 1
				break
			}
		}
	}

	//fmt.Println("MAMA", perfs)
	//fmt.Println("PAPA", userPerf)

	return
}

func calculatePerformance(fastestOfAll int, highestSuccessRate int, userFastest int, userSuccessRate int) int {
	// higher is better

	fastestRank := float64(userFastest) / float64(fastestOfAll) * 1000           // 1000 if you're the best, 1 if you're not the best
	successRank := float64(userSuccessRate) / float64(highestSuccessRate) * 1000 // 1000 if you're the best, 1 if you're not the best

	return int(successRank + fastestRank/2.0)
}
