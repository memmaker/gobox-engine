package ui

import (
	"testing"

	"gobox/game"
)

// TestHallKeepsWalls: the training hall (a Trading state) stays over the 3D
// view it was entered from; a shop does not.
func TestHallKeepsWalls(t *testing.T) {
	_, kept := wallsOf(game.Exploring{Map: game.ExploreMap{Walls: game.MapWalls{Side: [][][4]int{{{1, 0, 0, 0}}}}}}, nil)
	if kept == nil {
		t.Fatal("exploring kept no walls")
	}
	if shown, _ := wallsOf(game.Trading{Kind: game.TradeTraining}, kept); shown != kept {
		t.Error("the training hall dropped the 3D view")
	}
	if shown, _ := wallsOf(game.Trading{Kind: game.TradeShop}, kept); shown != nil {
		t.Error("a shop kept the 3D view")
	}
}
