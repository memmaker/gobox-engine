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
