package present

import (
	"testing"

	"gobox/combat"
	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// settled is g's Settle: the View after it and the moments queued as the
// presenter plays them.
func settled(g *Game) (front.View, []frame) {
	ms := g.Settle()
	v := g.View()
	var p Presenter
	p.Play(ms)
	return v, p.replay.q
}

// presenterOf is a Presenter over g.
func presenterOf(g *Game) *Presenter {
	p := &Presenter{}
	p.SetEngine(&classicGame{Game: g, p: p})
	return p
}

// boardIn is c as the tabletop draws it, over no game.
func boardIn(c *play.Combat, party []game.Member) *combat.Board {
	return new(Presenter).boardOf(c, party, nil)
}

// liveBoard is g's fight as the presenter draws it now (nil: none).
func liveBoard(g *Game) *combat.Board {
	p := presenterOf(g)
	p.fightView = g.combat()
	return p.aimPreview(p.boardOf(p.fightView, g.party(), nil))
}

// boardOfFrame is the board frame f holds, as drawn.
func boardOfFrame(f frame) *combat.Board { return new(Presenter).boardOf(f.Board, f.Party, f.fly) }

// played is es as moments (g.moments) queued as the presenter plays them.
func played(es []play.Report, fight bool, posed func(play.Report) (*play.Combat, []game.Member), holdMs int) []frame {
	if posed == nil {
		posed = func(play.Report) (*play.Combat, []game.Member) { return nil, nil }
	}
	g := NewGame(&fake{st: play.Exploring{}}, game.Data{})
	var p Presenter
	p.Play(g.moments(es, fight, posed, holdMs))
	return p.replay.q
}

// Class 20b: play.Report kinds pass one to one; an EvAttack swing stays a
// swing (never a miss) and EvHit + EvDamage stay one hit and one damage.
func TestEventsFold(t *testing.T) {
	es := []play.Report{{Kind: game.EvAttack, Unit: 1, Target: 2}, {Kind: game.EvHit, Unit: 1, Target: 2},
		{Kind: game.EvDamage, Unit: 2, Target: 2, N: 5}, {Kind: game.EvAttack, Unit: 2, Target: 1}, {Kind: game.EvMiss, Unit: 2, Target: 1}}
	want := []game.EventKind{game.EvAttack, game.EvHit, game.EvDamage, game.EvAttack, game.EvMiss}
	out := played(es, false, nil, 0)
	if len(out) != len(want) {
		t.Fatalf("%d events, want %d", len(out), len(want))
	}
	for i, k := range want {
		if out[i].Kind != k {
			t.Errorf("event %d: %v, want %v", i, out[i].Kind, k)
		}
	}
	if out[2].N != 5 || out[2].Unit != 2 {
		t.Errorf("damage %+v", out[2])
	}
}
