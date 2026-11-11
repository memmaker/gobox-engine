package game

import "gobox/core"

// Point buy's table is core's (core.PointBuyCost); these take the stats in a
// game's order for the old per-game paths.
// ponytail: goes once every game's PointBuy is core.RollPointBuy (por-go rules, B8).
const (
	PointBuyBudget = int(core.PointBuyBudget)
	PointBuyMin    = int(core.PointBuyMin)
	PointBuyMax    = int(core.PointBuyMax)
)

// PointBuyCost is the points a final score costs (-1: outside the range).
func PointBuyCost(score int) int { return int(core.PointBuyCost(int32(score))) }

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
