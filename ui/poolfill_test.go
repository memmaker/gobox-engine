package ui

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// The party pack's fill counts the pack (PoolItems), not the member's readied gear too.
func TestPoolFillCountsPack(t *testing.T) {
	in := upOf(play.Exploring{Common: play.Common{PartyCap: 60, PoolItems: 10, Party: []game.Member{{Items: make([]game.Item, 14)}}}}, game.ItemsOwn)
	if in.used != 10 || in.capacity != 60 {
		t.Fatalf("fill %d/%d, want 10/60", in.used, in.capacity)
	}
}
