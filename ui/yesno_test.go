package ui

import (
	"testing"

	"gobox/game"
)

// Yes and No reach a question's replies as Reply{1}, Reply{0}.
func TestYesNoReplies(t *testing.T) {
	for k, want := range map[ActKind]int{ActYes: 1, ActNo: 0} {
		r := &fake{st: game.Scene{Common: game.Common{Ask: &game.Question{YesNo: true}}}}
		NewGame(r, game.Data{}).Answer(Action{Kind: k})
		if len(r.sent) != 1 || r.sent[0] != (game.Reply{N: want}) {
			t.Errorf("%v: %#v", k, r.sent)
		}
	}
}

// A Refusal's Yes/No (Attack Ally): Yes does the Yes option's Do; its No
// without a Do takes nothing, never Answer 0.
func TestRefusalYesNo(t *testing.T) {
	opts := []game.Option{{Label: "Yes", Do: game.Search{}}, {Label: "No"}}
	r := &fake{st: game.Exploring{}}
	g := NewGame(r, game.Data{})
	g.r.refusalAsk = &game.Ask{Kind: game.YesNo, Options: opts}
	g.Answer(Action{Kind: ActYes})
	if len(r.sent) != 1 || r.sent[0] != (game.Search{}) {
		t.Errorf("Yes: %#v", r.sent)
	}
	for _, k := range []ActKind{ActNo, ActOK} {
		r = &fake{st: game.Exploring{}}
		g = NewGame(r, game.Data{})
		g.r.refusalAsk = &game.Ask{Kind: game.YesNo, Options: opts}
		g.Answer(Action{Kind: k, N: map[ActKind]int{ActOK: 1}[k]})
		if len(r.sent) != 0 || g.pending() != nil {
			t.Errorf("%v on No: %#v", k, r.sent)
		}
	}
}
