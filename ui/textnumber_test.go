package ui

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestTypedNumberN: a typed amount is the answer's N too, its Text kept.
func TestTypedNumberN(t *testing.T) {
	r := &fake{st: play.Scene{Common: play.Common{Ask: &play.Question{Number: true}}}}
	NewGame(r, game.Data{}).Answer(Action{Kind: ActText, Text: "100"})
	if len(r.sent) != 1 || r.sent[0] != (play.Reply{N: 100, Text: "100"}) {
		t.Fatalf("sent %+v", r.sent)
	}
}
