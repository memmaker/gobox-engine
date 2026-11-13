package ui

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestLostUnitPicture: a unit whose picture does not load (a DAX block the
// file lacks) still stands as a figure, never as an empty square.
func TestLostUnitPicture(t *testing.T) {
	lost := game.Picture{File: "NOSUCH", Block: 6}
	c := &play.Combat{Units: []game.Unit{{ID: 7, Member: -1, Foe: true, Art: lost}, {ID: 8, Member: -1, Foe: true, Art: lost}}}
	for _, u := range boardIn(c, nil).Units {
		if u.Icon == nil {
			t.Fatalf("unit %d has no figure", u.ID)
		}
	}
}
