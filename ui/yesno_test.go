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
// without a Do takes nothing, never Answer 0. Enter (ActOK) names neither:
// it takes nothing and the question stays.
func TestRefusalYesNo(t *testing.T) {
	opts := []game.Option{{Label: "Yes", Do: game.Search{}}, {Label: "No"}}
	r := &fake{st: game.Exploring{}}
	g := NewGame(r, game.Data{})
	g.r.refusalAsk = &game.Ask{Kind: game.YesNo, Options: opts}
	g.Answer(Action{Kind: ActYes})
	if len(r.sent) != 1 || r.sent[0] != (game.Search{}) {
		t.Errorf("Yes: %#v", r.sent)
	}
	for k, stays := range map[ActKind]bool{ActNo: false, ActOK: true} {
		r = &fake{st: game.Exploring{}}
		g = NewGame(r, game.Data{})
		g.r.refusalAsk = &game.Ask{Kind: game.YesNo, Options: opts}
		g.Answer(Action{Kind: k})
		if len(r.sent) != 0 || (g.pending() != nil) != stays {
			t.Errorf("%v: sent %#v, question up %v, want %v", k, r.sent, g.pending() != nil, stays)
		}
	}
}

// Enter or "any key" (ActOK) never chooses: on a question with several
// options or a yes/no it sends nothing; where there is nothing to choose it
// goes on (bug class: a key took the first class of the hall's undrawn pick).
func TestOKNeverChooses(t *testing.T) {
	for name, tc := range map[string]struct {
		q    *game.Question
		want int
	}{
		"choice":   {&game.Question{Text: "Pick New Class:", Replies: []string{"Cleric", "Magic-User"}}, 0},
		"yes/no":   {&game.Question{YesNo: true, Replies: []string{"No", "Yes"}}, 0},
		"continue": {&game.Question{Continue: true}, 1},
		"one way":  {&game.Question{Replies: []string{"Press Enter"}}, 1},
	} {
		r := &fake{st: game.Scene{Common: game.Common{Ask: tc.q}}}
		NewGame(r, game.Data{}).Answer(Action{Kind: ActOK})
		if len(r.sent) != tc.want {
			t.Errorf("%s: sent %#v, want %d", name, r.sent, tc.want)
		}
	}
}
