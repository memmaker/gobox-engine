package present

import (
	"testing"

	"gobox/game"
)

// TestRosterPickerFits: a roster of names in narrow cells (POD's 11 by 6)
// lays each name beside its portrait inside the box, every portrait shown.
func TestRosterPickerFits(t *testing.T) {
	names := []string{"ABC", "ALDRIC", "CORWYN", "PIP", "SYLVAINE", "THORVALD"}
	s := shown{Picker: &game.Picker{Pics: make([]game.Picture, len(names)), Labels: names}}
	s.Picker.Cell.X, s.Picker.Cell.Y = 11, 6
	v := Picker(s.Picker, s.Cells, 0, 80, 25)
	for i, r := range v.Cells {
		if end := r.X1 + 12 + len(names[i]) - 1; end > r.X2 || end > 80-3 {
			t.Errorf("%s ends at column %d, its cell at %d", names[i], end, r.X2)
		}
		if v.Portraits[i] == nil {
			t.Errorf("%s has no portrait", names[i])
		}
	}
}

// TestPickerOnShownPage: asked for a wider page than the game shows (it
// keeps the original's 40 by 25), the picker lays out on the page shown.
func TestPickerOnShownPage(t *testing.T) {
	names := []string{"ABC", "ALDRIC", "CORWYN"}
	s := shown{Picker: &game.Picker{Pics: make([]game.Picture, len(names)), Labels: names}}
	s.Picker.Cell.X, s.Picker.Cell.Y = 11, 6
	for _, r := range Picker(s.Picker, s.Cells, 0, 53, 33).Cells {
		if r.X2 > 40-2 || r.Y2 > 25-2 {
			t.Fatalf("cell %+v off the 40x25 page", r)
		}
	}
}
