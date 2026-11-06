package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// The items page carries the member's published AC and THAC0.
func TestInventoryStats(t *testing.T) {
	s := game.State{Party: []game.Member{{AC: -5, THAC0: 7}}}
	if n := Inventory(&front.Inventory{}, s); n.AC != -5 || n.THAC0 != 7 || !n.Stats {
		t.Fatalf("%+v", n)
	}
}
