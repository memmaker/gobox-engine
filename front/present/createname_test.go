package present

import (
	"testing"

	"gobox/front"
	"gobox/game/play"
)

// TestCreateNameTyped: the name asked over the creation sheet takes the
// typed letters and sends the whole line once, on Enter.
func TestCreateNameTyped(t *testing.T) {
	p := draftAsked(&play.Draft{}, &play.Question{Text: "Character name:", Typed: true, Max: 15})
	for _, k := range []uint16{'B', 'u', 'c', 'k', '\r'} {
		p.Key(k)
	}
	if a := p.inv.acts; len(a) != 1 || a[0].Kind != front.ActText || a[0].Text != "Buck" {
		t.Fatalf("sent %+v", a)
	}
}
