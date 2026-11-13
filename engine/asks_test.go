package engine

import (
	"gobox/game"
	"testing"
)

// BUG-CLASSES 6: POR has no item act that leaves a question of its own
// (Drop and Sell ask nothing; gobox confirms). Trade, its one asking act in
// the original, is gone with the party's one inventory: it is refused, never
// taken and dropped.
func TestTradeRefused(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	if _, err := e.doNow(game.Give{From: 0, To: 1, Item: 0}); err == nil { // one party inventory: nothing to give
		t.Fatal("Give: no refusal")
	}
}
