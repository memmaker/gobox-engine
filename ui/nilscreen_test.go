package ui

import (
	"testing"

	"gobox/game/play"
)

// TestNoBlankPlaceMode: a place's state (camp .. loot) with a Question
// published draws its screen or that Question (the training hall's: its
// party menu, as SotSB loads into the hall), never a blank frame.
func TestNoBlankPlaceMode(t *testing.T) {
	q := &play.Question{Text: "x", Replies: []string{"Train Character", "Exit"}}
	c := play.Common{Ask: q}
	for _, st := range []play.State{play.Camping{Common: c}, play.Trading{Common: c}, play.Trading{Common: c, Kind: play.TradeTemple},
		play.Trading{Common: c, Kind: play.TradeTraining}, play.Trading{Common: c, Kind: play.TradeVault}, play.Looting{Common: c}} {
		if p, _ := presenterOn(st); !p.mainMenuUp() && p.askUp() == nil && !p.screenUp() && !p.inventoryUp() {
			t.Errorf("%T %+v: blank", st, st)
		}
	}
	p, _ := presenterOn(play.Trading{Kind: play.TradeTraining, Common: play.Common{Ask: &play.Question{Replies: []string{"Train Character"}}}})
	if m := p.mainMenu(); m.Open == nil || !m.Hall {
		t.Errorf("hall: menu %v hall %v", m.Open, m.Hall)
	}
}
