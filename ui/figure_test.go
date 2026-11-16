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

// TestBoardBigIcon: a unit's picture larger than a cell reaches the board at its own size (the original blits the block as loaded).
func TestBoardBigIcon(t *testing.T) {
	art := game.Picture{W: 24, H: 48, Pixels: make([]uint8, 24*48), Clear: -1}
	c := &game.Combat{Units: []game.Unit{{ID: 1, Foe: true, Art: art, Size: 2}}}
	b := (&Presenter{}).boardOf(c, nil, nil)
	if got := b.Units[0].IconSize; got.X != 24 || got.Y != 48 {
		t.Errorf("icon size %v, want 24x48", got)
	}
}
