package rules1e

import (
	"slices"
	"testing"
)

// TestValuatePinned pins the gem and jewel values (ovr22:19B2) for a fixed
// seed: the dice come in this order, so a change of order changes them.
func TestValuatePinned(t *testing.T) {
	s := &RuleState{RandSeed: 12345}
	s.Purse.Gems, s.Purse.Jewelry = 4, 4
	var got []uint16
	for _, jewel := range []bool{false, true, false, true, false, true, false, true} {
		it, ok := s.valuate(jewel)
		if !ok {
			t.Fatal("nothing to value")
		}
		got = append(got, it.Price)
	}
	want := []uint16{100, 1225, 500, 1687, 50, 4120, 1000, 5000} // the same as before the port (por randitem_engine)
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if _, ok := s.valuate(false); ok {
		t.Fatal("valued a gem the purse lacks")
	}
}
