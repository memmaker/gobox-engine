package classic

import (
	"image"
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

// TestViewMemberSelected: the temple's View (-1) views the selected member.
func TestViewMemberSelected(t *testing.T) {
	r := &seqRules{}
	g := NewGame(r, game.Data{})
	g.ViewMember(-1)
	if len(r.got) != 1 || r.got[0] != (game.ViewMember{Member: r.State().Selected}) {
		t.Fatalf("%#v", r.got)
	}
}

// TestAimAnswersInFight: a Square ask in a fight (ModeCombat, BR1's Target)
// takes the clicked square as its Answer, not as an Aim the ask refuses.
func TestAimAnswersInFight(t *testing.T) {
	g := &Game{R: &loreRules{s: game.State{Mode: game.ModeCombat, Pending: &game.Ask{Kind: game.Square}}}}
	if d, _ := g.decision(front.Action{Kind: front.ActAim, X: 14, Y: 8}); d != (game.Answer{At: image.Pt(14, 8)}) {
		t.Fatalf("%#v", d)
	}
}
