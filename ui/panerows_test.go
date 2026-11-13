package ui

import (
	"testing"

	"gobox/game"
)

// A picture with Frames and FrameMs shows one of them; without, its own Frame.
func TestLayerFrames(t *testing.T) {
	p, _ := presenterOn(game.Scene{Layers: []game.Layer{{Picture: game.Picture{Frame: 9}}, {Picture: game.Picture{Frame: 9}, Frames: []int{3, 4}, FrameMs: 100}}})
	ps := p.pictures()
	if f := ps[1].Frame; ps[0].Frame != 9 || f != 3 && f != 4 {
		t.Fatalf("%+v", ps)
	}
}
