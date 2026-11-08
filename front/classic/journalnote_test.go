package classic

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

func TestAddedToJournal(t *testing.T) {
	r := play.Exploring{Common: play.Common{Lore: []game.Lore{{Title: "Old"}}}}
	f := &fake{st: r}
	g := NewGame(f, game.Data{})
	if v := g.View(); v.Note.Text != "" {
		t.Fatalf("a loaded journal flashed %q", v.Note.Text)
	}
	r.Lore = append([]game.Lore{{Title: "Journal Entry 37"}}, r.Lore...)
	f.st = r
	if v := g.View(); v.Note.Text != "Added to journal" {
		t.Fatalf("new note: %q", v.Note.Text)
	}
}
