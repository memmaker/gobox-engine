package classic

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/game"
)

// endRules is a fight that one Quick ends: its Result carries the fight's
// events and the state after it is the result page (no Combat).
type endRules struct {
	seqRules
	over bool
}

func (r *endRules) State() game.State {
	if r.over {
		return game.State{Mode: game.ModeStory}
	}
	return game.State{Mode: game.ModeCombat, Combat: &game.Combat{Size: image.Pt(4, 4), Units: []game.Unit{{ID: 1}, {ID: 2, Foe: true}}}}
}
func (r *endRules) Apply(d game.Decision) game.Result {
	r.over = true
	return game.Result{Events: []game.Event{{Kind: game.EvAttack, Unit: 1, Target: 2, Ms: 300}, {Kind: game.EvDown, Unit: 2, Ms: 400}}}
}

// TestEndingEventsPlay: the events of a Result that ends the fight still
// play, each held on the fight as it stood, before the page after it.
func TestEndingEventsPlay(t *testing.T) {
	r := &endRules{}
	g := NewGame(r, game.Data{})
	g.Settle() // the fight shown
	g.Answer(front.Action{Kind: front.ActQuick})
	v := g.Settle()
	if v.Battle != nil || len(v.Events) != 2 {
		t.Fatalf("battle %v, %d events", v.Battle != nil, len(v.Events))
	}
	for _, e := range v.Events {
		if e.Battle == nil || e.Ms == 0 {
			t.Errorf("event %v not held on the fight (battle %v, ms %d)", e.Kind, e.Battle != nil, e.Ms)
		}
	}
}

// walkFallRules end the fight with a walk, a fall and a shot.
type walkFallRules struct{ endRules }

func (r *walkFallRules) Apply(d game.Decision) game.Result {
	r.over = true
	return game.Result{Events: []game.Event{{Kind: game.EvMove, Unit: 1, At: image.Pt(0, 0), To: image.Pt(2, 1)},
		{Kind: game.EvAttack, Unit: 1, Target: 2, Ms: 300}, {Kind: game.EvDown, Unit: 2, Ms: 400},
		{Kind: game.EvMissile, Unit: 1, At: image.Pt(2, 1), To: image.Pt(3, 3), Ms: 300}}}
}

// TestEndingEventsMove: the ending events play on the board the ones before
// them left: the walker stands where it walked, the fallen is gone.
func TestEndingEventsMove(t *testing.T) {
	r := &walkFallRules{}
	g := NewGame(r, game.Data{})
	g.Settle()
	g.Answer(front.Action{Kind: front.ActQuick})
	es := g.Settle().Events
	if len(es) != 4 || es[1].Battle == nil || es[3].Battle == nil {
		t.Fatalf("%d events", len(es))
	}
	unit := func(b *front.Battle, id int) *front.Unit {
		for k := range b.Units {
			if b.Units[k].ID == id {
				return &b.Units[k]
			}
		}
		return nil
	}
	if u := unit(es[1].Battle, 1); u == nil || u.X != 2 || u.Y != 1 {
		t.Errorf("walker %+v, want at 2,1", u)
	}
	if unit(es[2].Battle, 2) == nil {
		t.Error("the fall's own moment has no faller")
	}
	if u := unit(es[3].Battle, 2); u != nil {
		t.Errorf("fallen still on the board: %+v", u)
	}
	if es[3].Kind != front.EvMissile || es[3].To != image.Pt(3, 3) {
		t.Errorf("shot %+v", es[3])
	}
}
