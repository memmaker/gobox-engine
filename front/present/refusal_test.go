package present

import (
	"errors"
	"image"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// TestRefusalNotes: a refusal from any call (Buy from a shop's list) is the
// toast every screen draws, never silent.
func TestRefusalNotes(t *testing.T) {
	g := NewGame(&fake{st: play.Trading{}, err: errors.New("Not enough Money.")}, game.Data{})
	if g.Buy(0) == nil || g.r.toast.Text != "Not enough Money." {
		t.Fatalf("note %q", g.r.toast.Text)
	}
}

// refusing refuses move with a question put instead.
type refusing struct {
	fake
	ask func(play.Action) error
}

func (r *refusing) Do(a play.Action) (play.Outcome, error) {
	r.sent = append(r.sent, a)
	return play.Outcome{}, r.ask(a)
}

// TestOffMapStepAsks: a Move step off the map that the game refuses with an
// Ask (POR: "Flee:") puts that Ask; its first option sends its Do.
func TestOffMapStepAsks(t *testing.T) {
	r := &refusing{ask: func(a play.Action) error {
		if _, ok := a.(play.MoveTo); ok {
			return &play.Refusal{Ask: &play.Ask{Prompt: "Flee:", Options: []play.Option{{Label: "Yes", Do: play.Flee{}}, {Label: "No"}}}}
		}
		return nil
	}}
	g := NewGame(r, game.Data{})
	g.Answer(front.Action{Kind: front.ActWalk, X: -1, Y: 3})
	if len(r.sent) != 1 || r.sent[0] != (play.MoveTo{To: image.Pt(-1, 3)}) {
		t.Fatalf("sent %+v", r.sent)
	}
	if q := g.pending(); g.mode() != game.ModeAsk || q == nil || q.Prompt != "Flee:" {
		t.Fatalf("shown %v %+v", g.mode(), q)
	}
	g.Answer(front.Action{Kind: front.ActChoose, N: 0})
	if len(r.sent) != 2 || r.sent[1] != (play.Flee{}) {
		t.Fatalf("answer sent %+v", r.sent)
	}
}

// TestFleeRefusedAsText: a refusal's text in a fight is a held message too.
func TestLootLeaveRefused(t *testing.T) {
	g := NewGame(&fake{st: play.Looting{}, err: &play.Refusal{Text: "Leave the treasure?"}}, game.Data{})
	var no front.ConfirmNo
	if err := g.LootLeave(false); !errors.As(err, &no) {
		t.Fatalf("text refusal: %v", err)
	}
}
