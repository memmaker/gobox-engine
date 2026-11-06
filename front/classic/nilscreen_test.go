package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// TestNoBlankPlaceMode: a place's mode (camp .. loot) with its screen not up
// yet and a Question published draws that Question (the training hall's:
// its party menu, as SotSB loads into the hall), never a blank frame.
func TestNoBlankPlaceMode(t *testing.T) {
	for m := game.ModeCamp; m <= game.ModeLoot; m++ {
		s := game.State{Mode: m, Pending: &game.Ask{Kind: game.Choice, Prompt: "x", Options: []game.Option{{Label: "Train Character"}, {Label: "Exit"}}}}
		if v := View(front.View{}, s); v.MainMenu == nil && v.Ask == nil {
			t.Errorf("mode %d: blank", m)
		}
	}
	s := game.State{Mode: game.ModeTraining, Pending: &game.Ask{Kind: game.Choice, Options: []game.Option{{Label: "Train Character"}}}}
	if v := View(front.View{}, s); v.MainMenu == nil || !v.Hall {
		t.Errorf("hall: menu %v hall %v", v.MainMenu, v.Hall)
	}
}
