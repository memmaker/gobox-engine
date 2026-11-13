package core

import "testing"

// TestRNGIsTurboPascals: the first values of gobox/rules.Random from seed 1.
func TestRNGIsTurboPascals(t *testing.T) {
	r := RNG{State: 1}
	r.N(0)
	if r.State != 0x08088405+1 {
		t.Fatal(r.State)
	}
	got := r.N(100)
	if want := int(uint16(r.State>>16) % 100); got != want {
		t.Fatal(got, want)
	}
}
