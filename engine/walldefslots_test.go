package engine

import (
	"testing"

	"gobox/rules"
)

// TestWalldefFillsSlots: New Phlan's LoadWalls(Reset, ...) loads set 3's
// block 0, three wall sets, into slots 1..3; the slots after the first no
// longer keep the area left's blocks (back from Sokol Keep by boat, New
// Phlan's walls went missing).
func TestWalldefFillsSlots(t *testing.T) {
	e, err := NewPool("../data/por", "../data/por/sources")
	if err != nil {
		t.Skip(err)
	}
	const newPhlanSet, sokolSlot2, sokolSlot3 = 3, 20, 21
	e.set = newPhlanSet
	e.mapPairA = [rules.WallSlots]int{1, sokolSlot2, sokolSlot3}
	e.LoadWalldef(0, 1)
	want := [rules.WallSlots]wallChunk{{0, 1}, {0, 2}, {0, 3}}
	if got := e.wallChunks(); got != want {
		t.Errorf("slots hold %v, want %v", got, want)
	}
}
