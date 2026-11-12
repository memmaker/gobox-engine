package pool

import (
	"gobox/core"
	"testing"
)

// TestEmptyPartyIDStable: the empty party's placeholder (selected) has one
// CharID, nil has NoChar and NoItem: naming them registers no new entry.
func TestEmptyPartyIDStable(t *testing.T) {
	e := testEngine(t)
	e.state().Party = nil
	a, b := e.selected(), e.selected()
	if a != b {
		t.Errorf("placeholder ids %d, %d: a new entry per call", a, b)
	}
	if e.charID(nil) != core.NoChar || e.itemHandle(nil) != core.NoItem {
		t.Error("nil got an id")
	}
}
