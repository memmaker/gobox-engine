package present

import (
	"testing"

	"gobox/front"
)

// TestCreateNameTyped: the name asked over the creation sheet takes the
// typed letters and sends the whole line once, on Enter.
func TestCreateNameTyped(t *testing.T) {
	v := front.View{Creation: &front.Creation{}, Ask: &front.Ask{Kind: front.AskText, Max: 15, Menu: front.Menu{Prompt: "Character name:"}}}
	p := &Presenter{}
	for _, k := range []uint16{'B', 'u', 'c', 'k', '\r'} {
		p.Key(k, v)
	}
	if a := p.inv.acts; len(a) != 1 || a[0].Kind != front.ActText || a[0].Text != "Buck" {
		t.Fatalf("sent %+v", a)
	}
}
