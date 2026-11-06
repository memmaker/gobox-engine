package classic

import (
	"errors"
	"testing"

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
