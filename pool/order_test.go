package pool

import (
	"gobox/core"
	"rules1e"
	"testing"
)

// A card dragged onto another's place moves the member there, while the
// Party Order bar is up and only then.
func TestOrderDrag(t *testing.T) {
	e := &Engine{}
	e.initVM()
	a, b, c := &rules1e.Character{}, &rules1e.Character{}, &rules1e.Character{}
	e.state().Party = []*rules1e.Character{a, b, c}
	if e.state().MoveMember(0, 2) == nil || e.state().Party[0] != a {
		t.Fatal("moved without the bar")
	}
	e.setWorldWord(core.WordCampOn, 1)
	if e.state().MoveMember(0, 2) != nil || e.state().Party[0] != b || e.state().Party[1] != c || e.state().Party[2] != a {
		t.Fatalf("first to last: %v", e.state().Party)
	}
	if e.state().MoveMember(1, 3) == nil {
		t.Fatal("moved out of the party")
	}
}
