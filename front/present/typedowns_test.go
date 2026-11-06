package present

import (
	"testing"

	"gobox/front"
)

// TestTypedOwnsKeys: a typed question owns the keyboard over any screen:
// the letters reach its line even with the party menu's screen state up
// (the name after creation's Keep, over the sheet the menu opened).
func TestTypedOwnsKeys(t *testing.T) {
	ask := &front.Ask{Kind: front.AskText, Max: 15, Menu: front.Menu{Prompt: "Character name:"}}
	for name, v := range map[string]front.View{
		"creation":  {Creation: &front.Creation{}, Ask: ask},
		"main menu": {Creation: &front.Creation{}, MainMenu: []front.ActKind{front.ActCreate}, Ask: ask},
	} {
		p := &Presenter{main: true}
		for _, k := range []uint16{'B', 'o', 'b', '\r'} {
			p.Key(k, v)
		}
		if a := p.inv.acts; len(a) != 1 || a[0].Kind != front.ActText || a[0].Text != "Bob" {
			t.Errorf("%s: sent %+v", name, a)
		}
	}
}
