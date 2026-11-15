package ui

import (
	"testing"

	"gobox/game"
)

// TestPartyQuestionShown: a question over the party menu that is not the
// menu itself (the hall's "Pick New Class:") is asked, never a blank frame
// whose Enter answers it unseen.
func TestPartyQuestionShown(t *testing.T) {
	g, f := slotsGame(t)
	f.st = game.Assembling{Common: game.Common{Ask: &game.Question{Text: "Pick New Class:", Replies: []string{"Cleric", "Magic-User"}}}}
	p := &Presenter{}
	p.SetEngine(g)
	if a := p.askUp(); a == nil || a.Kind != AskMenu || a.Prompt != "Pick New Class:" || len(a.Commands) != 2 {
		t.Fatalf("the party screen's question: %+v", a)
	}
	p.backdrop = true // the question on no screen: no key answers it
	p.Key(13)
	p.Key('M')
	if as := p.Actions(); len(as) != 0 || !p.unseen() {
		t.Fatalf("keys on an unseen question: %+v (unseen %v)", as, p.unseen())
	}
}
