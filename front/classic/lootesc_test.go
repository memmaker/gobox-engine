package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// TestLootFirstEsc: Esc on the found-treasure message leaves, as Enter then
// Esc does; it was swallowed.
func TestLootFirstEsc(t *testing.T) {
	r := &fake{st: play.Looting{Common: allow(play.Leave{}), Items: []game.Item{{Name: "Gauntlets"}}}}
	g := NewGame(r, game.Data{})
	g.View()
	g.Answer(front.Action{Kind: front.ActClose})
	if len(r.sent) != 1 || r.sent[0] != (play.Leave{}) {
		t.Fatalf("first Esc sent %v", r.sent)
	}
}
