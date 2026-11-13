package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// TestPowersShow: the powers a game offers on a member reach the inventory.
func TestPowersShow(t *testing.T) {
	st := play.Camping{Common: play.Common{Party: []game.Member{{Powers: []game.PowerOffer{{Label: "Heal", Kind: game.PowerHeal}, {Label: "Cure", Kind: game.PowerCure}}}}}}
	g := NewGame(&fake{st: st}, game.Data{})
	g.r.openItems(0)
	in := presenterOf(g).inventory()
	if in == nil || len(in.powers) != 2 || in.powers[0].Kind != front.ActHeal || in.powers[1].Kind != front.ActCure {
		t.Fatalf("%+v", in)
	}
}
