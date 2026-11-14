package ui

import (
	"image"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestOverlandGoesEachWay: on the overland the arrows are ActGo, a step
// north, east, south or west whatever the facing (the 3D view's forward and
// turns do not apply there), and the screen is the map even with the page's
// own small map in its layers, never a picture window (por 31bbc79, d9bc3f9).
func TestOverlandGoesEachWay(t *testing.T) {
	x := game.Exploring{Map: game.ExploreMap{Walk: yard{size: image.Pt(9, 9)}, Overland: true}}
	x.Where = game.Where{X: 4, Y: 4, Facing: 6}
	x.Layers = []game.Drawn{{}}
	if w := windowOf(x); w != game.WindowMap {
		t.Fatalf("overland window %v, want WindowMap", w)
	}
	for dir, to := range []rules.MapPos{{X: 4, Y: 3}, {X: 5, Y: 4}, {X: 4, Y: 5}, {X: 3, Y: 4}} {
		r := &fake{st: x}
		if err := NewGame(r, game.Data{}).Answer(Action{Kind: ActGo, N: dir}); err != nil {
			t.Fatal(err)
		}
		if len(r.sent) != 1 || r.sent[0] != (game.Arrive{From: rules.MapPos{X: 4, Y: 4}, To: to, Facing: 6}) {
			t.Errorf("ActGo %d sent %v, want Arrive to %v", dir, r.sent, to)
		}
	}
}

// TestMarkerFollowsArt: the party's marker is the art's current picture, not
// the first one seen (art without the marker cached an empty marker for good).
func TestMarkerFollowsArt(t *testing.T) {
	var a areaPic
	if a.marker("marker/0", map[string]game.Picture{}) != nil {
		t.Fatal("a marker from no art")
	}
	art := map[string]game.Picture{"marker/0": {W: 1, H: 1, Pixels: []byte{4}, Clear: -1}}
	if a.marker("marker/0", art) == nil {
		t.Error("the marker stays gone once the art has it")
	}
}
