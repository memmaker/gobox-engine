package ui

import (
	"testing"

	"gobox/game"
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
