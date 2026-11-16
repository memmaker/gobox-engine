package ui

import (
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestFigureActionPose: a unit in its action pose shows its Action picture when it has one, else its Art (as before).
func TestFigureActionPose(t *testing.T) {
	ready := game.Picture{W: 1, H: 1, Pixels: []uint8{1}}
	action := game.Picture{W: 1, H: 1, Pixels: []uint8{2}}
	for _, c := range []struct {
		name string
		u    game.Unit
		want uint8
	}{
		{"ready", game.Unit{Art: ready, Action: action, Foe: true}, 1},
		{"acting", game.Unit{Art: ready, Action: action, Pose: 1, Foe: true}, 2},
		{"acting without an action pose", game.Unit{Art: ready, Pose: 1, Foe: true}, 1},
	} {
		icon, _ := figureOf(c.u, nil)
		if len(icon.Pix) != 1 || icon.Pix[0] != c.want {
			t.Errorf("%s: pixels %v, want %d", c.name, icon.Pix, c.want)
		}
	}
}

// TestBoardOverlaySlots: two overlays of one size and different pixels (a cloud and a corpse) are two pictures on the board.
func TestBoardOverlaySlots(t *testing.T) {
	cloud := game.Picture{W: 2, H: 1, Pixels: []uint8{1, 2}, Clear: -1}
	body := game.Picture{W: 2, H: 1, Pixels: []uint8{3, 4}, Clear: -1}
	c := &game.Combat{Squares: []game.Ground{{At: rules.BoardPos{X: 1, Y: 1}, Overlay: cloud}, {At: rules.BoardPos{X: 2, Y: 1}, Overlay: body}}}
	b := (&Presenter{}).boardOf(c, nil, nil)
	a, z := b.Ground[1][1].Overlay, b.Ground[1][2].Overlay
	if a == 0 || z == 0 || a == z {
		t.Errorf("overlay slots %d and %d, want two different ones", a, z)
	}
}
