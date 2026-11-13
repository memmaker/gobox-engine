package ui

import (
	"testing"

	"gobox/game"
)

// TestPaneLineToast: over the items the text pane is hidden: a line the
// decision adds to it (BR2's shop "Doc buys a ...") is the toast.
func TestPaneLineToast(t *testing.T) {
	r := &paneGame{}
	r.st = game.Trading{}
	g := NewGame(r, game.Data{})
	g.r.items = game.ItemsOwn
	if err := g.do(game.Buy{}); err != nil {
		t.Fatal(err)
	}
	if g.r.toast.Text != "Doc buys a Rope" {
		t.Fatalf("note %q", g.r.toast.Text)
	}
}

type paneGame struct{ fake }

func (r *paneGame) Do(game.Action) (game.Outcome, error) {
	r.st = game.Trading{Common: game.Common{Text: []string{"Doc buys a Rope"}}}
	return game.Outcome{}, nil
}
