package ui

import (
	"image"
	"testing"
	"time"

	"gobox/game"
)

// With Ask.View the View action on a highlighted member reaches the rules' ViewMember; without, it does not.
func TestAskView(t *testing.T) {
	for _, view := range []bool{true, false} {
		st := game.Scene{Common: game.Common{Ask: &game.Question{Replies: []string{"A", "B", "C"}}, Party: []game.Member{{}, {}, {}}}}
		if view {
			st.Allowed.Plain = []game.Action{game.ViewMember{Member: 2}}
		}
		f := &fake{st: st}
		NewGame(f, game.Data{}).Answer(Action{Kind: ActView, N: 2})
		if view != (len(f.sent) == 1 && f.sent[0] == game.Action(game.ViewMember{Member: 2})) {
			t.Errorf("View %v: %v", view, f.sent)
		}
	}
}

// TestTempleView: View in a temple or shop is ViewMember on the selected member.
func TestTempleView(t *testing.T) {
	for _, st := range []game.State{
		game.Trading{Kind: game.TradeTemple, Common: game.Common{Selected: 1}, Services: []game.Service{{}}},
		game.Trading{Kind: game.TradeShop, Common: game.Common{Selected: 1}},
	} {
		g := NewGame(&fake{st: st}, game.Data{})
		g.Answer(Action{Kind: ActMember, N: -1})
		if g.r.sheet != 1 { // no View the rules offer: gobox's sheet of the selected member
			t.Fatalf("%T: sheet %d", st, g.r.sheet)
		}
	}
}

// TestAimAnswersInFight: a lone aim in a fight (BR1's Target) takes the
// clicked square as its answer, not as an Aim the ask refuses.
func TestAimAnswersInFight(t *testing.T) {
	st := game.Fighting{Common: game.Common{Allowed: game.Allowed{Aimed: []game.Aimed{{Action: game.Shoot{}, Squares: []image.Point{{14, 8}}, Targets: 1}}}}}
	f := &fake{st: st}
	g := NewGame(f, game.Data{})
	g.Answer(Action{Kind: ActAim, X: 14, Y: 8})
	if len(f.sent) != 1 {
		t.Fatalf("%#v", f.sent)
	}
}

// A picked rest time reaches the game as Rest{Minutes}.
func TestRestMinutes(t *testing.T) {
	f := &fake{st: game.Camping{}}
	NewGame(f, game.Data{}).Answer(Action{Kind: ActRest, N: 2, X: 300})
	if len(f.sent) != 1 || f.sent[0] != game.Action(game.AdvanceTimeBy{D: 300 * time.Minute}) {
		t.Fatalf("%#v", f.sent)
	}
}
