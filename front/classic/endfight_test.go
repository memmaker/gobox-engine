package classic

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// endRules is a fight that one Quick ends: its Outcome carries the fight's
// events and the state after it is the result page (no fight).
type endRules struct {
	fake
	events []play.Event
}

func (r *endRules) Do(a play.Action) (play.Outcome, error) {
	r.st = play.Scene{}
	return play.Outcome{Events: r.events}, nil
}

func newEnd(es ...play.Event) *endRules {
	r := &endRules{events: es}
	r.st = play.Fighting{Common: allow(play.Quick{}), Field: play.Battlefield{Size: image.Pt(4, 4), Units: []play.Unit{{ID: 1, Member: -1}, {ID: 2, Member: -1, Foe: true}}}}
	return r
}

// TestEndingEventsPlay: the events of a Result that ends the fight still
// play, each held on the fight as it stood, before the page after it.
func TestEndingEventsPlay(t *testing.T) {
	r := newEnd(play.Attacked{From: 1, To: 2}, play.Fell{Unit: 2, Ms: 400})
	g := NewGame(r, game.Data{})
	g.Settle() // the fight shown
	g.Answer(front.Action{Kind: front.ActQuick})
	v := g.Settle()
	if v.Battle != nil || len(v.Events) != 3 {
		t.Fatalf("battle %v, %d events", v.Battle != nil, len(v.Events))
	}
	for _, e := range []front.CombatEvent{v.Events[0], v.Events[2]} { // the blow and the fall; a miss has no moment
		if e.Battle == nil {
			t.Errorf("event %v not held on the fight (battle %v, ms %d)", e.Kind, e.Battle != nil, e.Ms)
		}
	}
}

// TestEndingEventsMove: the ending events play on the board the ones before
// them left: the walker stands where it walked, the fallen is gone.
func TestEndingEventsMove(t *testing.T) {
	r := newEnd(play.Walked{Unit: 1, Path: []game.Step{{At: image.Pt(2, 1)}}}, play.Attacked{From: 1, To: 2}, play.Fell{Unit: 2, Ms: 400},
		play.Shot{From: 1, To: image.Pt(3, 3)})
	g := NewGame(r, game.Data{})
	g.Settle()
	g.Answer(front.Action{Kind: front.ActQuick})
	es := g.Settle().Events
	if len(es) != 5 || es[1].Battle == nil || es[4].Battle == nil {
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
	if unit(es[3].Battle, 2) == nil {
		t.Error("the fall's own moment has no faller")
	}
	if u := unit(es[4].Battle, 2); u != nil {
		t.Errorf("fallen still on the board: %+v", u)
	}
	if es[4].Kind != front.EvMissile || es[4].To != image.Pt(3, 3) {
		t.Errorf("shot %+v", es[4])
	}
}
