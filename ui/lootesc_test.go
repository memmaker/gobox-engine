package ui

import (
	"testing"

	"gobox/game"
)

// TestLootFirstEsc: Esc on the found-treasure message leaves, as Enter then
// Esc does; it was swallowed.
func TestLootFirstEsc(t *testing.T) {
	r := &fake{st: game.Looting{Common: allow(game.Leave{}), Items: []game.Item{{Name: "Gauntlets"}}, Announce: true}}
	g := NewGame(r, game.Data{})
	g.Answer(Action{Kind: ActClose})
	if len(r.sent) != 2 || r.sent[0] != (game.ContinueAction{}) || r.sent[1] != (game.Leave{}) {
		t.Fatalf("first Esc sent %v", r.sent)
	}
}
