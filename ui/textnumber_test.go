package ui

import (
	"testing"

	"gobox/game"
)

// TestTypedNumberN: a typed amount is the answer's N too, its Text kept.
func TestTypedNumberN(t *testing.T) {
	r := &fake{st: game.Scene{Common: game.Common{Ask: &game.Question{Number: true}}}}
	NewGame(r, game.Data{}).Answer(Action{Kind: ActText, Text: "100"})
	if len(r.sent) != 1 || r.sent[0] != (game.Reply{N: 100, Text: "100"}) {
		t.Fatalf("sent %+v", r.sent)
	}
}
