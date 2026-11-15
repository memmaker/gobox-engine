package ui

import (
	"testing"

	"gobox/game"
	"gobox/ui/input"
)

// Search looks at this square once (the original's Look); Sweep turns the
// searching pace on or off (the original's Search).
func TestSearchAndSweep(t *testing.T) {
	for k, want := range map[ActKind]game.Action{ActSearch: game.Look{}, ActSweep: game.Search{}} {
		r := &fake{st: game.Exploring{}}
		NewGame(r, game.Data{}).Answer(Action{Kind: k, N: -1})
		if len(r.sent) != 1 || r.sent[0] != want {
			t.Errorf("%v: %#v, want %#v", k, r.sent, want)
		}
	}
}

// A cursor key pages only a question up: with none, no ActPage goes to the game.
func TestArrowPagesOnlyAQuestion(t *testing.T) {
	p := &Presenter{}
	p.arrow(input.Down)
	if len(p.inv.acts) != 0 {
		t.Errorf("no question: %#v", p.inv.acts)
	}
	var got []Action
	p.q = &Question{Answer: func(a Action) bool { got = append(got, a); return false }}
	p.arrow(input.Down)
	if len(got) != 1 || got[0] != (Action{Kind: ActPage, N: 1}) {
		t.Errorf("question up: %#v", got)
	}
}
