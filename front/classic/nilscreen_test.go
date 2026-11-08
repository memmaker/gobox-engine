package classic

import (
	"testing"

	"gobox/game"
)

// TestNoBlankPlaceMode: a place's mode (camp .. loot) with its screen not up
// yet and a Question published draws that Question (the training hall's:
// its party menu, as SotSB loads into the hall), never a blank frame.
func TestNoBlankPlaceMode(t *testing.T) {
	for m := game.ModeCamp; m <= game.ModeLoot; m++ {
		s := shown{Mode: m, Pending: &game.Ask{Kind: game.Choice, Prompt: "x", Options: []game.Option{{Label: "Train Character"}, {Label: "Exit"}}}}
		if v := menusOf(s); v.MainMenu == nil && v.Ask == nil {
			t.Errorf("mode %d: blank", m)
		}
	}
	s := shown{Mode: game.ModeTraining, Pending: &game.Ask{Kind: game.Choice, Options: []game.Option{{Label: "Train Character"}}}}
	if v := menusOf(s); v.MainMenu == nil || !v.Hall {
		t.Errorf("hall: menu %v hall %v", v.MainMenu, v.Hall)
	}
}
