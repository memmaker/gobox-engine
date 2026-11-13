package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// TestPlaceYesNoDrawn: a yes/no the game asks over a place's screen (BR2's
// "go back and claim it?" over the treasure, a shop's drop) is the view's
// confirm, the screen staying, never a frame without it.
func TestPlaceYesNoDrawn(t *testing.T) {
	for _, m := range []game.Mode{game.ModeLoot, game.ModeShop, game.ModeCamp} {
		s := shown{Mode: m, Loot: &game.Loot{}, Pending: &play.Ask{Kind: game.YesNo, Prompt: "Do you want to go back and claim it?", Default: 1,
			Options: []play.Option{{Label: "Yes", Do: play.Reply{N: 1}}, {Label: "No", Do: play.Reply{N: 2}}}}}
		if v := menusOf(s); v.Ask == nil || v.Ask.Kind != front.AskConfirm {
			t.Errorf("mode %d: no confirm (%+v)", m, v.Ask)
		}
	}
}

// TestCreateNameAsked: the name the game asks after creation's Keep is the
// view's typed line over the sheet (BR2: letters went out as picks).
func TestCreateNameAsked(t *testing.T) {
	s := shown{Mode: game.ModeCreate, Creation: &game.Creation{}, Pending: &play.Ask{Kind: game.Name, Prompt: "Character name:", Max: 15}}
	if v := menusOf(s); v.Ask == nil || v.Ask.Kind != front.AskText {
		t.Fatalf("no typed line: %+v", v.Ask)
	}
}
