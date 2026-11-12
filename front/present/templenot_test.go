package present

import (
	"strings"
	"testing"

	"gobox/front"
)

// TestTempleNotDeclines: a service that helps no one asks "Cast cure anyway?", unless the game's
// temple only says why (Cure.Declines): then it is told, with nothing to agree to.
func TestTempleNotDeclines(t *testing.T) {
	for _, declines := range []bool{false, true} {
		p := &Presenter{}
		p.camp.screen, p.camp.pick = templeNot, 0
		v := front.View{Temple: &front.Temple{Cures: []front.Cure{{Name: "Cure Blindness", Not: "is not blind", Declines: declines}}}}
		text, rows := p.templeItems(v)
		if asks := strings.Contains(text, "Cast cure anyway?"); asks == declines {
			t.Errorf("Declines %v: text %q", declines, text)
		}
		if (len(rows) == 1) != declines {
			t.Errorf("Declines %v: %d rows", declines, len(rows))
		}
	}
}

// TestTempleWoundOff: a wound cure (Cure.Only) a member does not need is greyed out:
// its row has nothing to pick, for key, mouse and pad alike (ui.RowsOff); a needed one is picked.
func TestTempleWoundOff(t *testing.T) {
	p := &Presenter{}
	p.camp.screen = templeHeal
	v := front.View{Temple: &front.Temple{Cures: []front.Cure{{Name: "Cure Light Wounds", Only: true}, {Name: "Cure Light Wounds", Only: true, Needed: true}, {Name: "Cure Blindness"}}}}
	_, rows := p.templeItems(v)
	if rows[0].Do != nil || rows[1].Do == nil || rows[2].Do == nil {
		t.Errorf("rows picked: %v %v %v, want off, on, on", rows[0].Do != nil, rows[1].Do != nil, rows[2].Do != nil)
	}
}
