package present

import (
	"testing"

	"gobox/front"
)

// TestLostUnitPicture: a unit whose picture does not load (a DAX block the
// file lacks) still stands as a figure, never as an empty square.
func TestLostUnitPicture(t *testing.T) {
	b := &front.Battle{Units: []front.Unit{{ID: 7}, {ID: 8}}, IconPix: []front.Picture{{File: "NOSUCH", Block: 6}, {File: "NOSUCH", Block: 6}}}
	for _, u := range pixBattle(b, egaInk).Units {
		if u.Icon == nil {
			t.Fatalf("unit %d has no figure", u.ID)
		}
	}
}
