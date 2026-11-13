package ui

import (
	"testing"

	"gobox/game"
)

// TestCreateNameTyped: the name asked over the creation sheet takes the
// typed letters and sends the whole line once, on Enter.
func TestCreateNameTyped(t *testing.T) {
	p := draftAsked(&game.Draft{}, &game.Question{Text: "Character name:", Typed: true, Max: 15})
	for _, k := range []uint16{'B', 'u', 'c', 'k', '\r'} {
		p.Key(k)
	}
	if a := p.inv.acts; len(a) != 1 || a[0].Kind != ActText || a[0].Text != "Buck" {
		t.Fatalf("sent %+v", a)
	}
}
