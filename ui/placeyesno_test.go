package ui

import (
	"testing"

	"gobox/game"
)

// TestPlaceYesNoDrawn: a yes/no the game asks over a place's screen (BR2's
// "go back and claim it?" over the treasure, a shop's drop) is the
// confirm, the screen staying, never a frame without it.
func TestPlaceYesNoDrawn(t *testing.T) {
	c := game.Common{Ask: &game.Question{Text: "Do you want to go back and claim it?", YesNo: true, Replies: []string{"No", "Yes"}, Default: 1}}
	for _, st := range []game.State{game.Looting{Common: c, Items: []game.Item{{Name: "Sword"}}}, game.Trading{Common: c}, game.Camping{Common: c}} {
		if p, _ := presenterOn(st); p.askUp() == nil || p.askUp().Kind != AskConfirm {
			t.Errorf("%T: no confirm (%+v)", st, p.askUp())
		}
	}
}

// TestCreateNameAsked: the name the game asks after creation's Keep is the
// typed line over the sheet (BR2: letters went out as picks).
func TestCreateNameAsked(t *testing.T) {
	p := draftAsked(&game.Draft{}, &game.Question{Text: "Character name:", Typed: true, Max: 15})
	if a := p.askUp(); a == nil || a.Kind != AskText {
		t.Fatalf("no typed line: %+v", a)
	}
}
