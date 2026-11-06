package classic

import (
	"testing"

	"gobox/game"
)

// TestPaneLineToast: over the items the text pane is hidden: a line the
// decision adds to it (BR2's shop "Doc buys a ...") is the toast.
func TestPaneLineToast(t *testing.T) {
	r := &paneRules{s: game.State{Items: game.ItemsShop, Page: &game.Page{}}}
	g := &Game{R: r}
	if err := g.do(game.Buy{}); err != nil {
		t.Fatal(err)
	}
	if g.note.Text != "Doc buys a Rope" {
		t.Fatalf("note %q", g.note.Text)
	}
}

type paneRules struct {
	game.Rules
	s game.State
}

func (r *paneRules) State() game.State { return r.s }
func (r *paneRules) Apply(game.Decision) game.Result {
	r.s.Page = &game.Page{Pane: [][]game.TextRun{{{Text: "Doc buys a Rope"}}}}
	return game.Result{}
}
