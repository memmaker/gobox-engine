package present

import (
	"testing"

	"gobox/front"
	"gobox/game/play"
)

// TestTypedOwnsKeys: a typed question owns the keyboard over any screen:
// the letters reach its line even with the party menu's screen state up
// (the name after creation's Keep, over the sheet the menu opened).
func TestTypedOwnsKeys(t *testing.T) {
	ask := play.Common{Ask: &play.Question{Text: "Character name:", Typed: true, Max: 15}}
	for name, st := range map[string]play.State{
		"creation":  play.Assembling{Common: ask, Draft: &play.Draft{}},
		"main menu": play.Assembling{Common: ask},
	} {
		p, _ := presenterOn(st)
		p.main = true
		for _, k := range []uint16{'B', 'o', 'b', '\r'} {
			p.Key(k)
		}
		if a := p.inv.acts; len(a) != 1 || a[0].Kind != front.ActText || a[0].Text != "Bob" {
			t.Errorf("%s: sent %+v", name, a)
		}
	}
}
