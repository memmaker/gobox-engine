package parity

import (
	"reflect"
	"testing"

	"gobox/front"
	"gobox/front/classic"
)

// TestNotesParity: the status line's toast and the after-combat page
// classic takes from State equal por's, over the dev scenes, Enter on.
func TestNotesParity(t *testing.T) {
	seen := map[string]int{}
	for _, scene := range []string{"", "autocombat"} {
		e := engine(t, true, scene)
		settled(e, front.Action{})
		for i := 0; i < 3; i++ {
			v, s := e.View(), e.Rules().State()
			b := v
			b.Note, b.Results = front.Note{}, nil
			c := classic.Results(classic.Note(b, s), s)
			if c.Note != v.Note || !reflect.DeepEqual(c.Results, v.Results) {
				t.Errorf("%s: note %+v results %+v, want %+v %+v", scene, c.Note, c.Results, v.Note, v.Results)
			}
			if v.Note.Text != "" {
				seen["note"]++
			}
			if v.Results != nil {
				seen["results"]++
			}
			settled(e, front.Action{Kind: front.ActOK})
		}
	}
	t.Logf("compared: %v", seen)
	if seen["note"] == 0 { // no dev scene reaches the after-combat page: classic's own test covers Results
		t.Errorf("not seen: %v", seen)
	}
}
