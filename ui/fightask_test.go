package ui

import (
	"testing"

	"gobox/game"
)

// TestFightAsk: a question the game puts mid-fight (Common.Ask) shows over
// the combat map: both are up.
func TestFightAsk(t *testing.T) {
	fight := game.Fighting{Common: game.Common{Ask: &game.Question{Text: "VIEW", Replies: []string{"GEAR", "EXIT"}}},
		Field: game.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, Member: 0}}}}
	g := NewGame(&fake{st: fight}, game.Data{})
	if q := presenterOf(g).askUp(); q == nil || g.combat() == nil {
		t.Fatalf("ask %+v, battle %v", q, g.combat() != nil)
	}
}
