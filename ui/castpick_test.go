package ui

import (
	"testing"

	"gobox/game"
)

// powerCastGame refuses a power's cast as the engine does: the spell pick
// first, then the member pick; with both it takes the cast.
type powerCastGame struct {
	fake
	cast *game.UsePower
}

func (g *powerCastGame) Do(a game.Action) (game.Outcome, error) {
	u, ok := a.(game.UsePower)
	switch {
	case !ok:
		return g.fake.Do(a)
	case u.Spell == 0:
		return game.Outcome{}, &game.Refusal{Ask: &game.Ask{Kind: game.PickSpell, Prompt: "Which spell?",
			Options: []game.Option{{Label: "Cure Light Wounds", Do: game.UsePower{Member: u.Member, Kind: u.Kind, Spell: 7, Target: -1}}}}}
	case u.Target < 0:
		return game.Outcome{}, &game.Refusal{Ask: &game.Ask{Kind: game.PickMember, Prompt: "Cast Spell on whom"}}
	}
	g.cast = &u
	return game.Outcome{}, nil
}

// castOn is a presenter exploring with two members, its game powerCastGame.
func powerCastOn() (*Presenter, *powerCastGame) {
	g := &powerCastGame{fake: fake{st: game.Exploring{Common: game.Common{Party: []game.Member{{Name: "A"}, {Name: "B"}}}}}}
	p := &Presenter{}
	p.SetEngine(&classicGame{Game: NewGame(g, game.Data{}), p: p})
	p.active = true
	return p, g
}

// flush takes the actions the presenter queued (a question's answer,
// else a power) and makes the rule calls they queued, as a frame does.
func flush(p *Presenter) {
	for len(p.inv.acts) > 0 || len(p.calls) > 0 {
		cs := p.calls
		p.calls = nil
		for _, c := range cs {
			c(p.eng)
		}
		as := p.inv.acts
		p.inv.acts = nil
		for _, a := range as {
			if taken, _ := AskAct(p, a); !taken {
				PowerAct(p, a)
			}
		}
	}
}

// TestCardCastPicks: a card's Cast puts up the spell pick (even for one
// spell), then the member pick; F2 casts on member 2.
func TestCardCastPicks(t *testing.T) {
	p, g := powerCastOn()
	p.memberAct(ActSpells, 0)
	flush(p)
	if len(p.spellPick()) != 1 {
		t.Fatalf("spell pick %v, want Cure Light Wounds up", p.spellPick())
	}
	p.send(ActCast, 7)
	flush(p)
	if !p.choosing() {
		t.Fatal("no member pick after the spell")
	}
	p.Key(KeyF1 + 1<<8)
	flush(p)
	if g.cast == nil || g.cast.Spell != 7 || g.cast.Target != 1 {
		t.Errorf("cast %+v, want spell 7 on member 1", g.cast)
	}
}

// TestCardCastEsc: Esc and a right click on the member pick give the cast up.
func TestCardCastEsc(t *testing.T) {
	for _, right := range []bool{false, true} {
		p, g := powerCastOn()
		p.memberAct(ActSpells, 0)
		flush(p)
		p.send(ActCast, 7)
		flush(p)
		if right {
			p.Mouse(5, 5, true, false, true, 0)
		} else {
			p.Key(27)
		}
		flush(p)
		if p.choosing() || g.cast != nil {
			t.Errorf("right %v: still asking %v, cast %+v", right, p.choosing(), g.cast)
		}
	}
}
