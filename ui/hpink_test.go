package ui

import "testing"

// TestHPInk: hit points are green above two thirds of the most, yellow above
// one third, red below.
func TestHPInk(t *testing.T) {
	for _, c := range []struct{ hp, max, want int }{{30, 30, inkGreen}, {21, 30, inkGreen}, {20, 30, inkYellow}, {11, 30, inkYellow}, {10, 30, inkRed}, {0, 30, inkRed}} {
		if got := hpInk(c.hp, c.max); got != c.want {
			t.Errorf("%d of %d: ink %d, want %d", c.hp, c.max, got, c.want)
		}
	}
}
