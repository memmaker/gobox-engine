package ui

import (
	"testing"

	"gobox/game"
)

// TestHallKeepsWalls: every shop-style menu (a Trading state) stays over the
// 3D view it was entered from.
func TestHallKeepsWalls(t *testing.T) {
	_, kept := wallsOf(game.Exploring{Map: game.ExploreMap{Walls: game.MapWalls{Side: [][][4]int{{{1, 0, 0, 0}}}}}}, nil)
	if kept == nil {
		t.Fatal("exploring kept no walls")
	}
	for _, k := range []game.TradeKind{game.TradeShop, game.TradeTemple, game.TradeTraining, game.TradeVault} {
		if shown, _ := wallsOf(game.Trading{Kind: k}, kept); shown != kept {
			t.Errorf("trade kind %v dropped the 3D view", k)
		}
	}
}
