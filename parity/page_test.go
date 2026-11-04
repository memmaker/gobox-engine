package parity

import (
	"testing"

	"gobox/front"
)

// TestPageParity: State.Page (the pane, the prompt row, the view window's
// text and the pictures) at each stop of the dev scenes, Enter on.
func TestPageParity(t *testing.T) {
	seen := map[string]int{}
	for _, scene := range []string{"", "password", "parlay", "who", "journal", "talljournal", "runes", "lockeddoor", "shop", "temple", "appraise", "loot", "effects", "autocombat"} {
		e := engine(t, true, scene)
		settled(e, front.Action{})
		for i := 0; i < 12; i++ {
			at(t, scene, e)
			if p := e.Rules().State().Page; p != nil {
				for k, on := range map[string]bool{"pane": len(p.Pane) > 0, "prompt": len(p.Prompt) > 0, "view": len(p.View) > 0, "pictures": len(p.Layers) > 0} {
					if on {
						seen[k]++
					}
				}
			}
			settled(e, front.Action{Kind: front.ActOK})
		}
	}
	for _, k := range []string{"pane", "prompt", "view", "pictures"} {
		if seen[k] == 0 {
			t.Errorf("no %s seen: %v", k, seen)
		}
	}
	t.Logf("page parts seen: %v", seen)
}
