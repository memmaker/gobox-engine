package ui

import (
	"testing"

	"gobox/game"
)

// TestNoBlankPlaceMode: a place's state (camp .. loot) with a Question
// published draws its screen or that Question (the training hall's: its
// party menu, as SotSB loads into the hall), never a blank frame.
func TestNoBlankPlaceMode(t *testing.T) {
	q := &game.Question{Text: "x", Replies: []string{"Train Character", "Exit"}}
	c := game.Common{Ask: q}
	for _, st := range []game.State{game.Camping{Common: c}, game.Trading{Common: c}, game.Trading{Common: c, Kind: game.TradeTemple},
		game.Trading{Common: c, Kind: game.TradeTraining}, game.Trading{Common: c, Kind: game.TradeVault}, game.Looting{Common: c}} {
		if p, _ := presenterOn(st); !p.mainMenuUp() && p.askUp() == nil && !p.screenUp() && !p.inventoryUp() {
			t.Errorf("%T %+v: blank", st, st)
		}
	}
	p, _ := presenterOn(game.Trading{Kind: game.TradeTraining, Common: game.Common{Ask: &game.Question{Replies: []string{"Train Character"}}}})
	if m := p.mainMenu(); m.Open == nil || !m.Hall {
		t.Errorf("hall: menu %v hall %v", m.Open, m.Hall)
	}
}
