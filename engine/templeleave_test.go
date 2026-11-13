//go:build devviews

package engine

import (
	"testing"

	"gobox/game"
)

// TestTempleLeave: the temple's Exit (game.Leave, which its State allows)
// leaves it, as the shop's does; leave() had no case for PendTemple and
// refused it "not now", so the party could not walk out of a temple.
func TestTempleLeave(t *testing.T) {
	e := devPlay(t, "temple")
	if !e.waiting(PendTemple) {
		t.Fatal("no temple")
	}
	if _, err := e.doNow(game.Leave{}); err != nil {
		t.Fatal("Leave:", err)
	}
	if e.waiting(PendTemple) {
		t.Fatal("still in the temple")
	}
}
