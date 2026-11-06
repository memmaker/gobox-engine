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

type lootRules struct {
	seqRules
	got []game.Decision
}

func (r *lootRules) Apply(d game.Decision) game.Result {
	r.got = append(r.got, d)
	if l, ok := d.(game.LootLeave); ok && !l.Sure {
		return game.Result{Err: &game.Refusal{Ask: &game.Ask{Kind: game.YesNo, Prompt: "There is still treasure left.  Do you want to go back and claim your treasure?",
			Options: []game.Option{{Label: "Yes"}, {Label: "No", Do: game.LootLeave{Sure: true}}}}}}
	}
	return game.Result{}
}

// TestLootLeaveAsksGame: leaving treasure behind puts the game's own
// question (POR's): Yes stays on the loot, No leaves (its Do), never the
// reverse; a refusal with text alone is asked so too (ConfirmNo).
func TestLootLeaveAsksGame(t *testing.T) {
	r := &lootRules{}
	g := NewGame(r, game.Data{})
	if err := g.LootLeave(false); err != nil {
		t.Fatalf("err %v", err)
	}
	if s := g.Shown(); s.Pending == nil || s.Mode != game.ModeAsk {
		t.Fatal("no question")
	}
	g.Answer(front.Action{Kind: front.ActYes})
	if len(r.got) != 1 {
		t.Fatalf("Yes sent %v", r.got)
	}
	g.LootLeave(false)
	g.Answer(front.Action{Kind: front.ActNo})
	if len(r.got) != 3 || r.got[2] != (game.LootLeave{Sure: true}) {
		t.Fatalf("No sent %v", r.got)
	}
	g = NewGame(&textLoot{}, game.Data{})
	var no front.ConfirmNo
	if err := g.LootLeave(false); !errors.As(err, &no) {
		t.Fatalf("text refusal: %v", err)
	}
}

type textLoot struct{ seqRules }

func (r *textLoot) Apply(game.Decision) game.Result {
	return game.Result{Err: &game.Refusal{Text: "Leave the treasure?"}}
}
