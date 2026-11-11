package core

// Point buy is creation's third roll kind (RollPointBuy), the same in every
// game: every ability starts at PointBuyMin and the player spends up to
// PointBuyBudget points, at PointBuyCost per final score. The ruleset then
// applies a roll's rules to the scores bought (age, race and class limits,
// exceptional strength, hit points). Felix 2026-11-11.
const (
	PointBuyBudget int32 = 32
	PointBuyMin    int32 = 8
	PointBuyMax    int32 = 18
)

var pointBuyCosts = [PointBuyMax - PointBuyMin + 1]int32{0, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16}

// PointBuyCost is the points a final score costs (-1: outside the range).
func PointBuyCost(score int32) int32 {
	if score < PointBuyMin || score > PointBuyMax {
		return -1
	}
	return pointBuyCosts[score-PointBuyMin]
}

// PointBuySpent is the points scores cost together; ok is false when one
// is outside the range or they cost more than the budget.
func PointBuySpent(scores map[Ability]int32) (spent int32, ok bool) {
	for _, s := range scores {
		c := PointBuyCost(s)
		if c < 0 {
			return 0, false
		}
		spent += c
	}
	return spent, spent <= PointBuyBudget
}
