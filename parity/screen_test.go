package parity

import (
	"image/color"
	"testing"

	"gobox/front"
)

// TestScreenParity: State's page size, text colours and journal text, over
// the dev scenes, Enter on.
func TestScreenParity(t *testing.T) {
	seen := map[string]int{}
	for _, scene := range []string{"", "journal", "talljournal", "runes", "effects"} {
		e := engine(t, true, scene)
		settled(e, front.Action{})
		for i := 0; i < 6; i++ {
			at(t, scene, e)
			s := e.Rules().State()
			if s.Cells.X > 0 {
				seen["cells"]++
			}
			if s.Ink != ([16]color.RGBA{}) {
				seen["ink"]++
			}
			if s.Read != nil {
				seen["read"]++
			}
			settled(e, front.Action{Kind: front.ActOK})
		}
	}
	t.Logf("seen: %v", seen)
	if seen["cells"] == 0 || seen["ink"] == 0 {
		t.Errorf("not seen: %v", seen)
	}
}
