package ui

import (
	"testing"

	"gobox/game"
)

// TestYesNoByLabel: a yes/no whose replies read Yes, No: the front's yes
// is the reply labelled Yes (index 0), not the play convention's 1.
func TestYesNoByLabel(t *testing.T) {
	g := &fake{st: game.Exploring{Common: game.Common{Ask: &game.Question{YesNo: true, Replies: []string{"Yes", "No"}}}}}
	r := newRules(g, game.Data{})
	r.send(tAnswer{N: 1})
	if len(g.sent) != 1 || g.sent[0] != (game.Reply{N: 0}) {
		t.Fatalf("yes sent %v", g.sent)
	}
}

// TestSceneNoPick: a scene's replies carry NoPick and NoEsc to the ask.
func TestSceneNoPick(t *testing.T) {
	g := &fake{st: game.Scene{Kind: game.SceneTalk, Common: game.Common{Ask: &game.Question{Replies: []string{"Yes", "No"}, NoPick: true, NoEsc: true}}}}
	s := shownOf(newRules(g, game.Data{}))
	if s.Pending == nil || !s.Pending.NoPick || !s.Pending.NoEsc {
		t.Fatalf("ask %+v", s.Pending)
	}
}
