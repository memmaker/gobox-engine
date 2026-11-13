package engine

import "testing"

// TestLabelInksEveryKind: Data.LabelInks (what the map legend and squares
// index by label kind) has an entry for every kind the engine defines, for
// every game (cotab and pod gave none: the map crashed on any label).
func TestLabelInksEveryKind(t *testing.T) {
	var e Engine
	inks := e.Data().LabelInks
	for k, name := range labelKinds {
		if k >= len(inks) {
			t.Errorf("label kind %d (%q) has no colour in Data.LabelInks %v", k, name, inks)
		}
	}
}
