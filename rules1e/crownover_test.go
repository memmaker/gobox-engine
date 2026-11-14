package rules1e

import "testing"

// Only the generator's real pairs are crown over trunk: a lone tall plant
// over grass or a plant tuft is not.
func TestCrownOver(t *testing.T) {
	w := PorGroundGen.Wild
	for _, c := range []struct {
		top, bottom Ground
		want        bool
	}{
		{0x41, 0x40, true},          // tall plant Top over Bottom
		{0x1F + 2, 0x23 + 2, true},  // grove of size 2
		{0x1F + 2, 0x23 + 3, false}, // grove sizes differ
		{0x3C + 1, 0x17, false},     // lone tall plant over grass
		{0x3C + 1, 0x36 + 1, false}, // over a low plant
	} {
		if got := w.crownOver(c.top, c.bottom); got != c.want {
			t.Errorf("crownOver(%#x, %#x) = %v, want %v", c.top, c.bottom, got, c.want)
		}
	}
	if (*WildGen)(nil).crownOver(0x41, 0x40) {
		t.Error("no wild gen must pair nothing")
	}
}
