package ui

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

func TestAddedToJournal(t *testing.T) {
	r := play.Exploring{Common: play.Common{Lore: []game.Lore{{Title: "Old"}}}}
	f := &fake{st: r}
	g := NewGame(f, game.Data{})
	if g.Settle(); g.r.toast.Text != "" {
		t.Fatalf("a loaded journal flashed %q", g.r.toast.Text)
	}
	r.Lore = append([]game.Lore{{Title: "Journal Entry 37"}}, r.Lore...)
	f.st = r
	if g.Settle(); g.r.toast.Text != "Added to journal" {
		t.Fatalf("new note: %q", g.r.toast.Text)
	}
}
