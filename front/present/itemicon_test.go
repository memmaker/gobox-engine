package present

import (
	"testing"

	"gobox/front"
)

// Any pool game's item finds its icon by name, the longest match winning.
func TestItemIconByName(t *testing.T) {
	var inv inventory
	inv.load()
	sword, two := inv.icon(&front.InvItem{Name: "Long Sword +1"}), inv.icon(&front.InvItem{Name: "Two-Handed Sword"})
	if sword < 0 || two < 0 || sword == two || inv.icon(&front.InvItem{Name: "zzz"}) != -1 {
		t.Fatalf("icons: long %d, two-handed %d", sword, two)
	}
}
