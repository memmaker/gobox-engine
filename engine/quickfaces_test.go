package engine

import (
	"testing"

	"gobox/game"
)

// TestQuickPartyFaces: the Quick Start party's members show their
// templates' faces (Data.QuickParty Head, Body), also those the game stores
// as its own portrait numbers (asset.Set): before, those published no pick
// and the front drew a race-and-class fallback.
func TestQuickPartyFaces(t *testing.T) {
	e := testEngine(t)
	e.Open("")
	e.titleDone()
	e.Advance()
	if _, err := e.doNow(game.QuickStart{}); err != nil {
		t.Fatal(err)
	}
	party := game.CommonOf(e.State()).Party
	for i, q := range e.gameData.QuickParty {
		if i >= len(party) {
			break
		}
		if m := party[i]; m.PortraitHeadID != q.Head || m.PortraitBodyID != q.Body {
			t.Errorf("%s: head %q body %q, want %q %q", m.Name, m.PortraitHeadID, m.PortraitBodyID, q.Head, q.Body)
		}
	}
}
