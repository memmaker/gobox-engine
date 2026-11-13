package ui

import (
	"testing"

	"gobox/game"
)

// TestOwnMemberPickArrows: Down moves the presenter's own member pick (Trade
// to whom?) from its candidate, and Enter takes the one moved to.
func TestOwnMemberPickArrows(t *testing.T) {
	p := partyOn(game.Member{Name: "A"}, game.Member{Name: "B"}, game.Member{Name: "C"})
	p.pageMenu = true // over a page (the items)
	got := -1
	AskForMember(p, "Trade to whom?", 0, func(n int) { got = n })
	for _, k := range []uint16{keyDown, keyDown, '\r'} {
		p.Key(k)
		for _, a := range p.inv.acts {
			AskAct(p, a)
		}
		p.inv.acts = nil
	}
	if got != 2 {
		t.Fatalf("Down, Down, Enter picked member %d, want 2", got)
	}
}
