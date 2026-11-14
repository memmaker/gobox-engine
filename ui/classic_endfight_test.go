package ui

import (
	"testing"

	"gobox/game"
	"gobox/rules"
	"gobox/ui/combat"
)

// endRules is a fight that one Quick ends: its Outcome carries the fight's
// events and the state after it is the result page (no fight).
type endRules struct {
	fake
	events []game.Event
}

func (r *endRules) Do(a game.Action) (game.Outcome, error) {
	r.st = game.Scene{}
	return game.Outcome{Events: r.events}, nil
}

func newEnd(es ...game.Event) *endRules {
	r := &endRules{events: es}
	r.st = game.Fighting{Common: allow(game.Quick{}), Field: game.Battlefield{Size: rules.Size{W: 4, H: 4}, Units: []game.Unit{{ID: 1, Member: -1}, {ID: 2, Member: -1, Foe: true}}}}
	return r
}

// TestEndingEventsPlayClassic: the events of a Result that ends the fight still
// play, each held on the fight as it stood, before the page after it.
func TestEndingEventsPlayClassic(t *testing.T) {
	r := newEnd(game.Attacked{From: 1, To: 2}, game.Fell{Unit: 2})
	g := NewGame(r, game.Data{})
	g.Settle() // the fight shown
	g.Answer(Action{Kind: ActQuick})
	evs := settled(g)
	if g.combat() != nil || len(evs) != 3 {
		t.Fatalf("battle %v, %d events", g.combat() != nil, len(evs))
	}
	for _, e := range []frame{evs[0], evs[2]} { // the blow and the fall; a miss has no moment
		if boardOfFrame(e) == nil {
			t.Errorf("event %v not held on the fight (battle %v, ms %d)", e.Kind, boardOfFrame(e) != nil, e.Ms)
		}
	}
}

// TestEndingEventsMove: the ending events play on the board the ones before
// them left: the walker stands where it walked, the fallen is gone.
func TestEndingEventsMove(t *testing.T) {
	r := newEnd(game.Walked{Unit: 1, Path: []game.Step{{At: rules.BoardPos{X: 2, Y: 1}}}}, game.Attacked{From: 1, To: 2}, game.Fell{Unit: 2},
		game.Shot{From: 1, To: rules.BoardPos{X: 3, Y: 3}})
	g := NewGame(r, game.Data{})
	g.Settle()
	g.Answer(Action{Kind: ActQuick})
	es := settled(g)
	if len(es) != 5 || boardOfFrame(es[1]) == nil || boardOfFrame(es[4]) == nil {
		t.Fatalf("%d events", len(es))
	}
	unit := func(b *combat.Board, id int) *combat.Unit {
		for k := range b.Units {
			if b.Units[k].ID == id {
				return &b.Units[k]
			}
		}
		return nil
	}
	if u := unit(boardOfFrame(es[1]), 1); u == nil || u.X != 2 || u.Y != 1 {
		t.Errorf("walker %+v, want at 2,1", u)
	}
	if unit(boardOfFrame(es[3]), 2) == nil {
		t.Error("the fall's own moment has no faller")
	}
	if u := unit(boardOfFrame(es[4]), 2); u != nil {
		t.Errorf("fallen still on the board: %+v", u)
	}
	if es[4].Kind != game.EvMissile || es[4].To != (rules.BoardPos{X: 3, Y: 3}) {
		t.Errorf("shot %+v", es[4])
	}
}
