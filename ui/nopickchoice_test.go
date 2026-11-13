package ui

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
	"gobox/ui/input"
)

// TestNoPickChoice: NoPick and NoEsc hold on a choice as on a yes/no:
// Enter and Esc do nothing, the hotkey answers.
func TestNoPickChoice(t *testing.T) {
	p := askedOn(&play.Ask{Kind: game.Choice, NoPick: true, NoEsc: true, Options: []play.Option{{Label: "Yes", Key: 'Y'}, {Label: "No", Key: 'N'}}})
	for _, k := range []uint16{keyOf(input.Enter), keyOf(input.Esc), keyOf(input.Down)} {
		if !p.pageMenuKey(k) || len(p.inv.acts) != 0 {
			t.Fatalf("key %d answered: %v", k, p.inv.acts)
		}
	}
	p.pageMenuKey('n')
	if len(p.inv.acts) != 1 || p.inv.acts[0].Kind != ActChoose || p.inv.acts[0].N != 1 {
		t.Fatalf("N: %v", p.inv.acts)
	}
}
