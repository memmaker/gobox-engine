package ui

import (
	"testing"

	"gobox/game"
)

// templeState is a temple with these services up for one member.
func templeState(sv ...game.Service) game.Trading {
	return game.Trading{Kind: game.TradeTemple, Services: sv, Common: game.Common{Party: []game.Member{{Name: "Ann"}}}}
}

// TestTempleOff (Felix's ruling): a service that would not help the selected
// member is greyed out: its row has nothing to pick, for key, mouse and pad
// alike (ui.RowsOff); a needed one is picked.
func TestTempleOff(t *testing.T) {
	g, _ := playGame(templeState(game.Service{Name: "Cure Disease"}, game.Service{Name: "Cure Light Wounds", Needs: []int{0}}))
	p := &Presenter{}
	p.SetEngine(g)
	p.camp.screen = templeHeal
	_, rows := p.templeItems()
	if rows[0].Do != nil || rows[1].Do == nil {
		t.Errorf("rows picked: %v %v, want off, on", rows[0].Do != nil, rows[1].Do != nil)
	}
}
