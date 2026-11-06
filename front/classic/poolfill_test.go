package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// The party pack's fill counts the pack (PoolItems), not the member's readied gear too.
func TestPoolFillCountsPack(t *testing.T) {
	s := game.State{PartyCap: 60, PoolItems: 10, Party: []game.Member{{Items: make([]game.Item, 14)}}}
	if in := Inventory(&front.Inventory{}, s); in.Used != 10 || in.Cap != 60 {
		t.Fatalf("fill %d/%d, want 10/60", in.Used, in.Cap)
	}
}
