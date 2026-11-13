package present

import (
	"testing"

	"gobox/game/play"
)

// TestMenuOverLastScene: a shop's or temple's page that publishes no map
// keeps the 3D scene last shown to draw over, never black.
func TestMenuOverLastScene(t *testing.T) {
	var p Presenter
	g, f := playGame(play.Exploring{Map: play.ExploreMap{Walls: play.Walls{Side: [][][4]int{{{0, 0, 0, 0}}}}}})
	p.SetEngine(g)
	shown := p.scene()
	f.st = play.Trading{}
	if p.scene() != nil || p.behind != shown {
		t.Fatal("the temple's page lost the town's scene")
	}
}
