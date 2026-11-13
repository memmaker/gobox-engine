package ui

import (
	"testing"

	"gobox/game"
)

// The items page carries the member's published AC and THAC0.
func TestInventoryStats(t *testing.T) {
	n := upOf(game.Exploring{Common: game.Common{Party: []game.Member{{AC: -5, THAC0: 7}}}}, game.ItemsOwn)
	if n.ac != -5 || n.thac0 != 7 || !n.stats {
		t.Fatalf("%+v", n)
	}
}
