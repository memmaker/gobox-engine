package ui

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestLootFirstEsc: Esc on the found-treasure message leaves, as Enter then
// Esc does; it was swallowed.
func TestLootFirstEsc(t *testing.T) {
	r := &fake{st: play.Looting{Common: allow(play.Leave{}), Items: []game.Item{{Name: "Gauntlets"}}, Announce: true}}
	g := NewGame(r, game.Data{})
	g.Answer(FrontAction{Kind: ActClose})
	if len(r.sent) != 2 || r.sent[0] != (play.Continue{}) || r.sent[1] != (play.Leave{}) {
		t.Fatalf("first Esc sent %v", r.sent)
	}
}
