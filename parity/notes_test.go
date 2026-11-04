package parity

import (
	"testing"

	"gobox/front"
)

// TestNotesParity: State's toast and after-combat page, over the dev
// scenes, Enter on.
func TestNotesParity(t *testing.T) {
	seen := map[string]int{}
	for _, scene := range []string{"", "autocombat"} {
		e := engine(t, true, scene)
		settled(e, front.Action{})
		for i := 0; i < 3; i++ {
			at(t, scene, e)
			s := e.Rules().State()
			if s.Note.Text != "" {
				seen["note"]++
			}
			if s.Results != nil {
				seen["results"]++
			}
			settled(e, front.Action{Kind: front.ActOK})
		}
	}
	t.Logf("seen: %v", seen)
	if seen["note"] == 0 { // no dev scene reaches the after-combat page: classic's own test covers Results
		t.Errorf("not seen: %v", seen)
	}
}
