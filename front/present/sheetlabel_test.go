package present

import (
	"strings"
	"testing"

	"gobox/front"
	"gobox/gpu"
)

// TestSheetHidesEmptyLabels: a field the game leaves empty draws no label.
func TestSheetHidesEmptyLabels(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	var drawn strings.Builder
	g := func(c byte) [8]byte { drawn.WriteByte(c); return [8]byte{} }
	note := Note{Glyph: g, Font: func(c byte, _ bool) [8]byte { return g(c) }}
	v := front.View{Party: []front.Member{{Name: "A", HP: 5, MaxHP: 5, Selected: true}}, SheetPage: &front.SheetPage{}, Sheet: &front.Sheet{}}
	(&Presenter{bars: gpu.NewImageFromImage(barArt())}).drawSheet(dst, note, v)
	for _, l := range []string{"Exp", "Level", "Damage"} {
		if strings.Contains(drawn.String(), l) {
			t.Errorf("%q drawn for an empty field", l)
		}
	}
}
