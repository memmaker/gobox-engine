package ui

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestShopMessageNotes: a message over the items (the shop's "buys a ...")
// is the toast the inventory draws, never lost.
func TestShopMessageNotes(t *testing.T) {
	r := &fake{st: play.Trading{}, out: play.Outcome{Events: []play.Event{play.Narrated{Text: "Brann buys a Broad Sword "}}}}
	g := NewGame(r, game.Data{})
	g.r.items = game.ItemsOwn
	if g.Buy(0); g.r.toast.Text != "Brann buys a Broad Sword" {
		t.Fatalf("note %q", g.r.toast.Text)
	}
}
