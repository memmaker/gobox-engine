package ui

import (
	"gobox/adventure"
	"testing"

	"gobox/game"
)

// A script's layer with Frames and FrameMs shows one of them; a drawn picture its own Frame.
func TestLayerFrames(t *testing.T) {
	withScripts(t, &adventure.Sequence{Steps: []adventure.SequenceStep{{Kind: adventure.StepPicture,
		Layers: []adventure.Layer{{Art: adventure.Art{File: "F", Block: 1}, Frames: []int{3, 4}, FrameMs: 100}}}}}, nil)
	p, _ := presenterOn(game.Scene{Kind: game.SceneTitle, Intro: &game.Intro{Pictures: []int{1}}, Common: game.Common{Picture: &game.Picture{Frame: 9}}})
	ps := p.pictures()
	if f := ps[1].Frame; ps[0].Frame != 9 || f != 3 && f != 4 {
		t.Fatalf("%+v", ps)
	}
}
