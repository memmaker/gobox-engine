package classic

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// With Ask.View the View action on a highlighted member reaches the rules' ViewMember; without, it does not.
func TestAskView(t *testing.T) {
	for _, view := range []bool{true, false} {
		st := play.Scene{Common: play.Common{Ask: &play.Question{Replies: []string{"A", "B", "C"}}, Party: []play.Member{{}, {}, {}}}}
		if view {
			st.Allowed.Plain = []play.Action{play.ViewMember{Member: 2}}
		}
		f := &fake{st: st}
		d, ok := NewGame(f, game.Data{}).decision(front.Action{Kind: front.ActView, N: 2})
		if v, isView := d.(game.ViewMember); view != (ok && isView && v.Member == 2) {
			t.Errorf("View %v: %v %v", view, d, ok)
		}
	}
}

// TestTempleView: View in a temple or shop is ViewMember on the selected member.
func TestTempleView(t *testing.T) {
	for _, st := range []play.State{
		play.Trading{Kind: play.TradeTemple, Common: play.Common{Selected: 1}, Services: []game.Service{{}}},
		play.Trading{Kind: play.TradeShop, Common: play.Common{Selected: 1}},
	} {
		g := NewGame(&fake{st: st}, game.Data{})
		if d, _ := g.decision(front.Action{Kind: front.ActMember, N: -1}); d != (game.ViewMember{Member: 1}) {
			t.Fatalf("%T: %#v", st, d)
		}
	}
}

// TestAimAnswersInFight: a lone aim in a fight (BR1's Target) takes the
// clicked square as its answer, not as an Aim the ask refuses.
func TestAimAnswersInFight(t *testing.T) {
	st := play.Fighting{Common: play.Common{Allowed: play.Allowed{Aimed: []play.Aimed{{Action: play.Shoot{}, Squares: []image.Point{{14, 8}}, Targets: 1}}}}}
	g := NewGame(&fake{st: st}, game.Data{})
	if d, _ := g.decision(front.Action{Kind: front.ActAim, X: 14, Y: 8}); d != (game.Answer{At: image.Pt(14, 8)}) {
		t.Fatalf("%#v", d)
	}
}

// A picked rest time reaches the game as Rest{Minutes}.
func TestRestMinutes(t *testing.T) {
	g := NewGame(&fake{st: play.Camping{}}, game.Data{})
	if d, _ := g.decision(front.Action{Kind: front.ActRest, N: 2, X: 300}); d != (game.Rest{Minutes: 300}) {
		t.Fatalf("%#v", d)
	}
}
