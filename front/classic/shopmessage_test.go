package classic

import (
	"testing"

	"gobox/game"
)

type buyRules struct{ seqRules }

func (r *buyRules) State() game.State { return game.State{Items: game.ItemsShop} }
func (r *buyRules) Apply(game.Decision) game.Result {
	return game.Result{Events: []game.Event{{Kind: game.EvMessage, Text: "Brann buys a Broad Sword ", Wait: true}}}
}

// TestShopMessageNotes: a message over the items (the shop's "buys a ...")
// is the toast the inventory draws, never lost.
func TestShopMessageNotes(t *testing.T) {
	g := NewGame(&buyRules{}, game.Data{})
	if g.Buy(0); g.note.Text != "Brann buys a Broad Sword" {
		t.Fatalf("note %q", g.note.Text)
	}
}
