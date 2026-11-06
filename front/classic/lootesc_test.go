package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// TestLootFirstEsc: Esc on the found-treasure message leaves (Cancel),
// as Enter then Esc does; it was swallowed.
func TestLootFirstEsc(t *testing.T) {
	r := &fightRules{s: game.State{Mode: game.ModeLoot, Party: []game.Member{{}}, Loot: &game.Loot{Items: []game.Item{{Name: "Gauntlets"}}}}}
	g := &Game{R: r}
	g.ViewOf(r.s)
	g.Answer(front.Action{Kind: front.ActClose})
	if len(r.got) != 1 {
		t.Fatalf("first Esc sent %v", r.got)
	}
	if _, ok := r.got[0].(game.Cancel); !ok {
		t.Fatalf("first Esc sent %T", r.got[0])
	}
}
