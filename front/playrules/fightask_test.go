package playrules

import (
	"testing"

	"gobox/front/classic"
	"gobox/game"
	"gobox/game/play"
)

// TestFightAsk: a question the game puts mid-fight (Common.Ask) shows over
// the combat map: the View has both.
func TestFightAsk(t *testing.T) {
	fight := play.Fighting{Common: play.Common{Ask: &play.Question{Text: "VIEW", Replies: []string{"GEAR", "EXIT"}}},
		Field: play.Battlefield{Turn: 1, Units: []play.Unit{{ID: 1, Member: 0}}}}
	v := classic.NewGame(New(&fake{st: fight}, game.Data{}), game.Data{}).View()
	if v.Ask == nil || v.Battle == nil {
		t.Fatalf("ask %+v, battle %v", v.Ask, v.Battle != nil)
	}
}
