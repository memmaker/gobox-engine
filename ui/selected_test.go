package ui

import (
	"testing"

	"gobox/game"
)

// The marked card is the selected member's, else the candidate of a member pick.
func TestPartySelectedFromIndex(t *testing.T) {
	g, _ := playGame(game.Exploring{Common: game.Common{Selected: 1, Party: []game.Member{{Name: "A"}, {Name: "B"}}}})
	p := &Presenter{}
	p.SetEngine(g)
	if n := p.cardSelected(); n != 1 || p.selectedIndex() != 1 || p.selectedName() != "B" {
		t.Fatalf("selected %d, want member 1", n)
	}
	AskForMember(p, "Whom?", 0, func(int) {})
	if n := p.cardSelected(); n != 0 {
		t.Fatalf("candidate %d, want member 0 alone", n)
	}
}
