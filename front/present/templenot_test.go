package present

import (
	"strings"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// templeState is a temple with these services up for one member.
func templeState(sv ...game.Service) play.Trading {
	return play.Trading{Kind: play.TradeTemple, Services: sv, Common: play.Common{Party: []game.Member{{Name: "Ann"}}}}
}

// TestTempleNotDeclines: a service that helps no one asks "Cast cure anyway?", unless the game's
// temple only says why (Service.Declines): then it is told, with nothing to agree to.
func TestTempleNotDeclines(t *testing.T) {
	for _, declines := range []bool{false, true} {
		g, _ := playGame(templeState(game.Service{Name: "Cure Blindness", Not: map[int]string{0: "is not blind"}, Declines: declines}))
		p := &Presenter{}
		p.SetEngine(g)
		p.camp.screen, p.camp.pick = templeNot, 0
		text, rows := p.templeItems()
		if asks := strings.Contains(text, "Cast cure anyway?"); asks == declines {
			t.Errorf("Declines %v: text %q", declines, text)
		}
		if (len(rows) == 1) != declines {
			t.Errorf("Declines %v: %d rows", declines, len(rows))
		}
	}
}

// TestTempleWoundOff: a wound cure (Service.Only) a member does not need is greyed out:
// its row has nothing to pick, for key, mouse and pad alike (ui.RowsOff); a needed one is picked.
func TestTempleWoundOff(t *testing.T) {
	g, _ := playGame(templeState(game.Service{Name: "Cure Light Wounds", Only: true}, game.Service{Name: "Cure Light Wounds", Only: true, Needs: []int{0}}, game.Service{Name: "Cure Blindness", Needs: []int{1}}))
	p := &Presenter{}
	p.SetEngine(g)
	p.camp.screen = templeHeal
	_, rows := p.templeItems()
	if rows[0].Do != nil || rows[1].Do == nil || rows[2].Do == nil {
		t.Errorf("rows picked: %v %v %v, want off, on, on", rows[0].Do != nil, rows[1].Do != nil, rows[2].Do != nil)
	}
}
