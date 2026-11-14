package ui

import (
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestShotWithoutArt: a Shot the game gave no Art (COTAB) carries no sprite,
// so no pictureless missile is flown over the board.
func TestShotWithoutArt(t *testing.T) {
	es := eventsAt([]game.Event{game.Shot{From: 1, To: rules.BoardPos{X: 3, Y: 4}}, game.Shot{From: 1, Art: game.Picture{File: "SPRIT", Block: 2}}}, 0, nil)
	if len(es) != 2 || len(es[0].Sprite) != 0 || len(es[1].Sprite) != 1 {
		t.Fatalf("sprites %v", es)
	}
}
