package ui

import (
	"testing"

	"gobox/game"
)

// TestRosterPickerFits: a roster of names in narrow cells (POD's 11 by 6)
// lays each name beside its portrait inside the box, every portrait shown.
func TestRosterPickerFits(t *testing.T) {
	names := []string{"ABC", "ALDRIC", "CORWYN", "PIP", "SYLVAINE", "THORVALD"}
	s := shown{Picker: &game.Picker{Pics: make([]game.Picture, len(names)), Labels: names}}
	s.Picker.Pics[0] = game.Picture{W: 88, H: 48} // 11 by 6 cells
	v := Picker(s.Picker, 0, 80, 25)
	for i, r := range v.Cells {
		if end := r.X1 + 12 + len(names[i]) - 1; end > r.X2 || end > 80-3 {
			t.Errorf("%s ends at column %d, its cell at %d", names[i], end, r.X2)
		}
		if v.Portraits[i] == nil {
			t.Errorf("%s has no portrait", names[i])
		}
	}
}

// TestPickerUsesPage: a larger page holds more pictures a page, every
// cell on it.
func TestPickerUsesPage(t *testing.T) {
	k := &game.Picker{Pics: make([]game.Picture, 60)}
	k.Pics[0] = game.Picture{W: 88, H: 48} // 11 by 6 cells
	small, big := Picker(k, 0, 40, 25), Picker(k, 0, 120, 60)
	if big.PerPage <= small.PerPage {
		t.Errorf("%d a page on 120x60, %d on 40x25", big.PerPage, small.PerPage)
	}
	for _, r := range big.Cells {
		if r.X2 > 120-2 || r.Y2 > 60-2 {
			t.Fatalf("cell %+v off the 120x60 page", r)
		}
	}
}
