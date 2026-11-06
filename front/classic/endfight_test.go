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
