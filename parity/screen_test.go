package parity

import (
	"image"
	"image/color"
	"reflect"
	"testing"

	"gobox/front"
	"gobox/front/classic"
)

// TestScreenParity: the page's size, the text colours and the journal text
// classic takes from State equal por's, over the dev scenes, Enter on.
func TestScreenParity(t *testing.T) {
	seen := map[string]int{}
	for _, scene := range []string{"", "journal", "talljournal", "runes", "effects"} {
		e := engine(t, true, scene)
		settled(e, front.Action{})
		for i := 0; i < 6; i++ {
			v, s := e.View(), e.Rules().State()
			b := v
			b.Cells, b.Ink, b.Read = image.Point{}, [16]color.RGBA{}, nil
			c := classic.Read(classic.Screen(b, s), s)
			if c.Cells != v.Cells || c.Ink != v.Ink || !reflect.DeepEqual(c.Read, v.Read) {
				t.Errorf("%s: cells %v ink %v read %+v, want %v %v %+v", scene, c.Cells, c.Ink, c.Read, v.Cells, v.Ink, v.Read)
			}
			if v.Cells.X > 0 {
				seen["cells"]++
			}
			if v.Ink != ([16]color.RGBA{}) {
				seen["ink"]++
			}
			if v.Read != nil {
				seen["read"]++
			}
			settled(e, front.Action{Kind: front.ActOK})
		}
	}
	t.Logf("compared: %v", seen)
	if seen["cells"] == 0 || seen["ink"] == 0 {
		t.Errorf("not seen: %v", seen)
	}
}
