package present

import (
	"testing"

	"gobox/front"
)

// TestMenuOverLastScene: a shop's or temple's page that publishes no map
// keeps the 3D scene last shown to draw over, never black.
func TestMenuOverLastScene(t *testing.T) {
	var p Presenter
	w := &front.Walls{Key: "town", Side: [][][4]int{{{0, 0, 0, 0}}}}
	shown := p.scene(front.View{Walls: w, Exploring: true})
	if p.scene(front.View{Temple: &front.Temple{}}) != nil || p.behind != shown {
		t.Fatal("the temple's page lost the town's scene")
	}
}
