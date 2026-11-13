package present

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestFightAsk: a question the game puts mid-fight (Common.Ask) shows over
// the combat map: the View has both.
func TestFightAsk(t *testing.T) {
	fight := play.Fighting{Common: play.Common{Ask: &play.Question{Text: "VIEW", Replies: []string{"GEAR", "EXIT"}}},
		Field: play.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, Member: 0}}}}
	g := NewGame(&fake{st: fight}, game.Data{})
	if v := g.View(); v.Ask == nil || g.combat() == nil {
		t.Fatalf("ask %+v, battle %v", v.Ask, g.combat() != nil)
	}
}
