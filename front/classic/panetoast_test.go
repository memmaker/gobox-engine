package classic

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestPaneLineToast: over the items the text pane is hidden: a line the
// decision adds to it (BR2's shop "Doc buys a ...") is the toast.
func TestPaneLineToast(t *testing.T) {
	r := &paneGame{}
	r.st = play.Trading{}
	g := NewGame(r, game.Data{})
	g.r.items = game.ItemsOwn
	if err := g.do(play.Buy{}); err != nil {
		t.Fatal(err)
	}
	if g.note.Text != "Doc buys a Rope" {
		t.Fatalf("note %q", g.note.Text)
	}
}

type paneGame struct{ fake }

func (r *paneGame) Do(play.Action) (play.Outcome, error) {
	r.st = play.Trading{Common: play.Common{Text: []string{"Doc buys a Rope"}}}
	return play.Outcome{}, nil
}
