package ui

import (
	"testing"

	"gobox/game"
)

// TestBareGroundTextured: a fight's bare ground (a tile all keyed clear, as
// POR's outdoor grass) shows the series floor texture as it is,
// never a flat fill (it drew plain black under the overland fights).
func TestBareGroundTextured(t *testing.T) {
	var p Presenter
	b := p.boardOf(&game.Combat{}, nil, nil)
	if b.GroundTex == nil {
		t.Fatal("no ground texture")
	}
	if _, flat := flatColour(b.GroundTex); flat {
		t.Error("ground texture is a flat fill")
	}
}
