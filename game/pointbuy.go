package game

// Point buy is creation's alternative to rolling the abilities, offered
// when Data.PointBuy says the game takes PointBuy. Every ability starts
// at PointBuyMin, and the player spends up to PointBuyBudget points, at
// PointBuyCost per final score. The cost table is gobox's and the same in
// every game. The game applies its own rules to the scores bought (race,
// class minimums, exceptional strength), as it does after a roll.
const (
	PointBuyBudget = 32
	PointBuyMin    = 8
	PointBuyMax    = 18
)

var pointBuyCosts = [PointBuyMax - PointBuyMin + 1]int{0, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16}

// PointBuyCost is the points a final score costs (-1: outside the range).
func PointBuyCost(score int) int {
	if score < PointBuyMin || score > PointBuyMax {
		return -1
	}
	return pointBuyCosts[score-PointBuyMin]
}

// PointBuySpent is the points scores cost together; ok is false when one
// is outside the range or they cost more than the budget.
func PointBuySpent(scores []int) (spent int, ok bool) {
	for _, s := range scores {
		c := PointBuyCost(s)
		if c < 0 {
			return 0, false
		}
		spent += c
	}
	return spent, spent <= PointBuyBudget
}
