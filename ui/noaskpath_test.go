package ui

import (
	"testing"

	"gobox/game/play"
	"gobox/ui/input"
)

// TestNoPickEnterAnyPath: Enter and Esc through the real Key handler never
// answer a NoPick/NoEsc choice, over a scene's text too; the hotkey does.
func TestNoPickEnterAnyPath(t *testing.T) {
	ask := &play.Question{Replies: []string{"Yes", "No"}, Keys: []rune{'Y', 'N'}, NoPick: true, NoEsc: true}
	chest := []string{"A CHEST."}
	for i, st := range []play.State{play.Scene{Common: play.Common{Ask: ask}}, play.Scene{Common: play.Common{Ask: ask, Text: chest}},
		play.Exploring{Common: play.Common{Ask: ask, Text: chest}}} {
		p, _ := presenterOn(st)
		p.Key(keyOf(input.Enter))
		p.Key(keyOf(input.Esc))
		if len(p.inv.acts) != 0 {
			t.Fatalf("view %d: answered %v", i, p.inv.acts)
		}
		p.Key('n')
		if len(p.inv.acts) != 1 || p.inv.acts[0].Kind != ActChoose || p.inv.acts[0].N != 1 {
			t.Fatalf("N: %v", p.inv.acts)
		}
	}
}
