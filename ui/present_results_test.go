package ui

import (
	"testing"

	"gobox/game"
	"gobox/gpu"
)

// The after-combat page is drawn from the state's Results, not from a View.
func TestResultsDrawnFromState(t *testing.T) {
	dst := gpu.NewImage(640, 400)
	defer dst.Deallocate()
	full := func(byte) [8]byte { return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	note := Note{Glyph: full, Font: func(byte, bool) [8]byte { return full(0) }}
	g, _ := playGame(game.Scene{Kind: game.SceneResults, Results: &game.Results{Outcome: "The party has won.", Exp: "Each character receives 120"}})
	p := &Presenter{}
	p.SetEngine(g)
	p.drawResults(dst, note)
	if n := lit(dst); n == 0 {
		t.Fatal("nothing drawn")
	}
}
