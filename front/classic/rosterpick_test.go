package classic

import (
	"testing"

	"gobox/game"
)

// TestRosterPickerFits: a roster of names in narrow cells (POD's 11 by 6)
// lays each name beside its portrait inside the box, every portrait shown.
func TestRosterPickerFits(t *testing.T) {
	names := []string{"ABC", "ALDRIC", "CORWYN", "PIP", "SYLVAINE", "THORVALD"}
	s := game.State{Picker: &game.Picker{Pics: make([]game.Picture, len(names)), Labels: names}}
	s.Picker.Cell.X, s.Picker.Cell.Y = 11, 6
	v := Picker(s, 0, 80, 25)
	for i, r := range v.Cells {
		if end := r.X1 + 12 + len(names[i]) - 1; end > r.X2 || end > 80-3 {
			t.Errorf("%s ends at column %d, its cell at %d", names[i], end, r.X2)
		}
		if v.Portraits[i] == nil {
			t.Errorf("%s has no portrait", names[i])
		}
	}
}
