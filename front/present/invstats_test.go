package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// The items page carries the member's published AC and THAC0.
func TestInventoryStats(t *testing.T) {
	s := shown{Party: []game.Member{{AC: -5, THAC0: 7}}}
	if n := Inventory(&front.Inventory{}, s.Mode, s.Party, s.Selected, s.PartyCap, s.PoolItems, s.Purse, s.Allowed, s.Shop, s.Loot); n.AC != -5 || n.THAC0 != 7 || !n.Stats {
		t.Fatalf("%+v", n)
	}
}
