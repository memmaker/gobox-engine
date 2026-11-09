package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// Yes and No reach a question's replies as Reply{1}, Reply{0}.
func TestYesNoReplies(t *testing.T) {
	for k, want := range map[front.ActKind]int{front.ActYes: 1, front.ActNo: 0} {
		r := &fake{st: play.Scene{Common: play.Common{Ask: &play.Question{YesNo: true}}}}
		NewGame(r, game.Data{}).Answer(front.Action{Kind: k})
		if len(r.sent) != 1 || r.sent[0] != (play.Reply{N: want}) {
			t.Errorf("%v: %#v", k, r.sent)
		}
	}
}

// A Refusal's Yes/No (Attack Ally): Yes does the Yes option's Do; its No
// without a Do takes nothing, never Answer 0.
func TestRefusalYesNo(t *testing.T) {
	opts := []play.Option{{Label: "Yes", Do: play.Search{}}, {Label: "No"}}
	r := &fake{st: play.Exploring{}}
	g := NewGame(r, game.Data{})
	g.asked = &play.Ask{Kind: game.YesNo, Options: opts}
	g.Answer(front.Action{Kind: front.ActYes})
	if len(r.sent) != 1 || r.sent[0] != (play.Search{}) {
		t.Errorf("Yes: %#v", r.sent)
	}
	for _, k := range []front.ActKind{front.ActNo, front.ActOK} {
		r = &fake{st: play.Exploring{}}
		g = NewGame(r, game.Data{})
		g.asked = &play.Ask{Kind: game.YesNo, Options: opts}
		g.Answer(front.Action{Kind: k, N: map[front.ActKind]int{front.ActOK: 1}[k]})
		if len(r.sent) != 0 || g.pending() != nil {
			t.Errorf("%v on No: %#v", k, r.sent)
		}
	}
}
