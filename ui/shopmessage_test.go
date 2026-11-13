package ui

import (
	"testing"

	"gobox/game"
)

// TestShopMessageNotes: a message over the items (the shop's "buys a ...")
// is the toast the inventory draws, never lost.
func TestShopMessageNotes(t *testing.T) {
	r := &fake{st: game.Trading{}, out: game.Outcome{Events: []game.Event{game.Narrated{Text: "Brann buys a Broad Sword "}}}}
	g := NewGame(r, game.Data{})
	g.r.items = game.ItemsOwn
	if g.Buy(0); g.r.toast.Text != "Brann buys a Broad Sword" {
		t.Fatalf("note %q", g.r.toast.Text)
	}
}
