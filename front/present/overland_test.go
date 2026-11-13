package present

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// TestOverlandGoesEachWay: on the overland the arrows are ActGo, a step
// north, east, south or west whatever the facing (the 3D view's forward and
// turns do not apply there), and the screen is the map even with the page's
// own small map in its layers, never a picture window (por 31bbc79, d9bc3f9).
func TestOverlandGoesEachWay(t *testing.T) {
	x := play.Exploring{Map: play.ExploreMap{Walk: yard{size: image.Pt(9, 9)}, Overland: true}}
	x.Where = game.Where{X: 4, Y: 4, Facing: 6}
	x.Layers = []game.Layer{{}}
	if w := windowOf(x); w != game.WindowMap {
		t.Fatalf("overland window %v, want WindowMap", w)
	}
	for dir, to := range []image.Point{{4, 3}, {5, 4}, {4, 5}, {3, 4}} {
		r := &fake{st: x}
		if err := NewGame(r, game.Data{}).Answer(front.Action{Kind: front.ActGo, N: dir}); err != nil {
			t.Fatal(err)
		}
		if len(r.sent) != 1 || r.sent[0] != (play.Arrive{From: image.Pt(4, 4), To: to, Facing: 6}) {
			t.Errorf("ActGo %d sent %v, want Arrive to %v", dir, r.sent, to)
		}
	}
}
