package pool

import (
	"gobox/rules1e"
	"slices"
	"testing"
)

// TestBracersLook: bracers are on the armour's scale (50 is AC 10), so
// Bracers +4 read AC 6, not "AC 50 better".
func TestBracersLook(t *testing.T) {
	e := testEngine(t)
	e.state().Items = e.readItems()
	if e.state().Base(rules1e.ItemBracers).Protection != rules1e.AcGiven|rules1e.AcTen {
		t.Skip("bracers not in the item table")
	}
	l := e.state().ItemLines(&rules1e.Item{Type: rules1e.ItemBracers, Plus: 4}, nil)
	if !slices.Contains(l, "AC 6, as armor") {
		t.Errorf("Look %q, want AC 6, as armor", l)
	}
	if p := e.state().Protection(&rules1e.Item{Type: rules1e.ItemBracers, Plus: 4}); p != 4 {
		t.Errorf("protection %v, want 4", p)
	}
}
