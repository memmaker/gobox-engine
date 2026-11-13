package game

import "testing"

type key struct{}

func TestHeld(t *testing.T) {
	var h Held
	n := 0
	load := func() int { n++; return n }
	if HeldValue(&h, key{}, load) != 1 || HeldValue(&h, key{}, load) != 1 {
		t.Fatal("waiting: read more than once")
	}
	done := h.Change()
	if HeldValue(&h, key{}, load) != 2 || HeldValue(&h, key{}, load) != 3 {
		t.Fatal("inside a change: held")
	}
	done()
	if HeldValue(&h, key{}, load) != 4 || HeldValue(&h, key{}, load) != 4 {
		t.Fatal("after a change: not read again once")
	}
}
