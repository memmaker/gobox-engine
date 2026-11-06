package classic

import (
	"errors"
	"image"
	"testing"

	"gobox/front"

	"gobox/game"
)

type refuseRules struct{ seqRules }

func (r *refuseRules) Apply(game.Decision) game.Result {
	return game.Result{Err: errors.New("Not enough Money.")}
}

// TestRefusalNotes: a refusal from any call (Buy from a shop's list) is the
// toast every screen draws, never silent.
func TestRefusalNotes(t *testing.T) {
	g := NewGame(&refuseRules{}, game.Data{})
	if g.Buy(0) == nil || g.note.Text != "Not enough Money." {
		t.Fatalf("note %q", g.note.Text)
	}
}

type fleeRules struct {
	seqRules
	got []game.Decision
}

func (r *fleeRules) Apply(d game.Decision) game.Result {
	r.got = append(r.got, d)
	if _, ok := d.(game.Move); ok {
		return game.Result{Err: &game.Refusal{Ask: &game.Ask{Prompt: "Flee:", Options: []game.Option{{Label: "Yes", Do: game.Flee{}}, {Label: "No"}}}}}
	}
	return game.Result{}
}

// TestOffMapStepAsks: a Move step off the map that the game refuses with an
// Ask (POR: "Flee:") puts that Ask; its first option sends its Do.
func TestOffMapStepAsks(t *testing.T) {
	r := &fleeRules{}
	g := NewGame(r, game.Data{})
	g.Answer(front.Action{Kind: front.ActWalk, X: -1, Y: 3})
	if len(r.got) != 1 || r.got[0] != (game.Move{To: image.Pt(-1, 3)}) {
		t.Fatalf("sent %+v", r.got)
	}
	if s := g.Shown(); s.Mode != game.ModeAsk || s.Pending == nil || s.Pending.Prompt != "Flee:" {
		t.Fatalf("shown %v %+v", s.Mode, s.Pending)
	}
	g.Answer(front.Action{Kind: front.ActChoose, N: 0})
	if len(r.got) != 2 || r.got[1] != (game.Flee{}) {
		t.Fatalf("answer sent %+v", r.got)
	}
}
