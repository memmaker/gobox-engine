package adventure

import "testing"

// TestReserveFor: the area's symbol set picks the reserve's monster; past the table none.
func TestReserveFor(t *testing.T) {
	a := Reserve{BySet: []ReserveUnit{{Monster: 59, Icon: 59}, {Monster: 6, Icon: 6}}}
	if m, ok := a.For(1); !ok || m.Monster != 6 {
		t.Fatalf("screen 1: %v %v", m, ok)
	}
	if _, ok := a.For(2); ok {
		t.Fatal("screen 2 has a monster")
	}
}
