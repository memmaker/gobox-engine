package play

import "testing"

type key struct{}

func TestHeld(t *testing.T) {
	var h Held
	n := 0
	load := func() int { n++; return n }
	if Hold(&h, key{}, load) != 1 || Hold(&h, key{}, load) != 1 {
		t.Fatal("waiting: read more than once")
	}
	done := h.Change()
	if Hold(&h, key{}, load) != 2 || Hold(&h, key{}, load) != 3 {
		t.Fatal("inside a change: held")
	}
	done()
	if Hold(&h, key{}, load) != 4 || Hold(&h, key{}, load) != 4 {
		t.Fatal("after a change: not read again once")
	}
}
