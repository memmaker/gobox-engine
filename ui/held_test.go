package ui

import "testing"

type heldKey struct{}

func TestHeld(t *testing.T) {
	var h heldStore
	n := 0
	load := func() int { n++; return n }
	if heldValue(&h, heldKey{}, load) != 1 || heldValue(&h, heldKey{}, load) != 1 {
		t.Fatal("waiting: read more than once")
	}
	done := h.Change()
	if heldValue(&h, heldKey{}, load) != 2 || heldValue(&h, heldKey{}, load) != 3 {
		t.Fatal("inside a change: held")
	}
	done()
	if heldValue(&h, heldKey{}, load) != 4 || heldValue(&h, heldKey{}, load) != 4 {
		t.Fatal("after a change: not read again once")
	}
}
