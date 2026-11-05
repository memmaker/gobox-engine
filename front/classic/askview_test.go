package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// With Ask.View the View action on a highlighted member is ViewMember; without, it is not.
func TestAskView(t *testing.T) {
	for _, view := range []bool{true, false} {
		g := NewGame(&seqRules{}, game.Data{})
		g.asked = &game.Ask{Kind: game.PickMember, View: view}
		d, ok := g.decision(front.Action{Kind: front.ActView, N: 2})
		if v, isView := d.(game.ViewMember); view != (ok && isView && v.Member == 2) {
			t.Errorf("View %v: %v %v", view, d, ok)
		}
	}
}

// TestTempleView: View in a temple or shop is ViewMember on the selected member.
func TestTempleView(t *testing.T) {
	for _, md := range []game.Mode{game.ModeTemple, game.ModeShop} {
		g := &Game{R: &loreRules{s: game.State{Mode: md, Selected: 1}}}
		if d, _ := g.decision(front.Action{Kind: front.ActMember, N: -1}); d != (game.ViewMember{Member: 1}) {
			t.Fatalf("mode %d: %#v", md, d)
		}
	}
}
