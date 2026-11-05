package classic

import (
	"testing"

	"gobox/game"
)

type loreRules struct {
	game.Rules
	s game.State
}

func (r *loreRules) State() game.State { return r.s }

func TestAddedToJournal(t *testing.T) {
	r := &loreRules{s: game.State{Lore: []game.Lore{{Title: "Old"}}}}
	g := &Game{R: r}
	if v := g.View(); v.Note.Text != "" {
		t.Fatalf("a loaded journal flashed %q", v.Note.Text)
	}
	r.s.Lore = append([]game.Lore{{Title: "Journal Entry 37"}}, r.s.Lore...)
	if v := g.View(); v.Note.Text != "Added to journal" {
		t.Fatalf("new note: %q", v.Note.Text)
	}
}
