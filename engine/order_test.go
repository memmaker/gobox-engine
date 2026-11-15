package engine

import (
	"errors"
	"gobox/rules"
	"gobox/rules1e"
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
	e.state().Camp.On = true // the rules' own camp flag, as MoveMember reads it
	if e.state().MoveMember(0, 2) != nil || e.state().Party[0] != b || e.state().Party[1] != c || e.state().Party[2] != a {
		t.Fatalf("first to last: %v", e.state().Party)
	}
	if !errors.Is(e.state().MoveMember(1, 3), rules1e.ErrNoMove) {
		t.Fatal("moved out of the party")
	}
	// the camp's offered Order with its places unset (From == To) is a bad
	// move, not "not now" (the offered-then-refused class, TestOfferedTaken)
	if err := e.state().MoveMember(0, 0); errors.Is(err, rules.ErrNotNow) || err == nil {
		t.Fatalf("Order{0,0}: %v, want ErrNoMove", err)
	}
}
