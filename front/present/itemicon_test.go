package present

import (
	"testing"
)

// Any pool game's item finds its icon by name, the longest match winning.
func TestItemIconByName(t *testing.T) {
	var inv inventory
	inv.load()
	sword, two := inv.icon("Long Sword +1", nil), inv.icon("Two-Handed Sword", nil)
	if sword < 0 || two < 0 || sword == two || inv.icon("zzz", nil) != -1 {
		t.Fatalf("icons: long %d, two-handed %d", sword, two)
	}
}
