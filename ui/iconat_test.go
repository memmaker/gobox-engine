package ui

import (
	"image"
	"testing"

	"gobox/rules"
)

// TestIconAtSizes: every series' item icon comes out at exactly the size asked, whatever its atlas cell (Pool's 20,
// Buck Rogers' 64), the one scaling the inventory, identify pop-up, camp menu and battle bar use. Detects the class: a
// screen sizing atlas cells by its own rule (the bar's 3x put a 64 pixel cell at 192 in a 64 pixel face).
func TestIconAtSizes(t *testing.T) {
	for _, s := range []string{"pool", "buck-rogers"} {
		inv := &inventory{}
		old := series
		series = s
		inv.load()
		series = old
		var k rules.ItemID
		for base := range inv.icons {
			k = base
			break
		}
		for _, n := range []int{iconSlot, 2 * iconSlot, 3 * 8, 60, 80} {
			ic := inv.iconAt(k, n)
			if ic == nil || ic.Bounds() != image.Rect(0, 0, n, n) {
				t.Fatalf("%s: item %d at %d: %v", s, k, n, ic)
			}
		}
	}
}
