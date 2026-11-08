package classic

import (
	"image"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestShotWithoutArt: a Shot the game gave no Art (COTAB) carries no sprite,
// so no pictureless missile is flown over the board.
func TestShotWithoutArt(t *testing.T) {
	es := eventsAt([]play.Event{play.Shot{From: 1, To: image.Pt(3, 4)}, play.Shot{From: 1, Art: game.Picture{File: "SPRIT", Block: 2}}}, 0, nil)
	if len(es) != 2 || len(es[0].Sprite) != 0 || len(es[1].Sprite) != 1 {
		t.Fatalf("sprites %v", es)
	}
}
