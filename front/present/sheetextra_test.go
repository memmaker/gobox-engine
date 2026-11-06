package present

import (
	"strings"
	"testing"

	"gobox/front"
	"gobox/gpu"
)

// TestSheetDrawsExtraFields: the two-panel sheet (shop, temple) draws the
// weapon, armour and load the game fills, as the page layout does.
func TestSheetDrawsExtraFields(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	var drawn strings.Builder
	g := func(c byte) [8]byte { drawn.WriteByte(c); return [8]byte{} }
	note := Note{Glyph: g, Font: func(c byte, _ bool) [8]byte { return g(c) }}
	v := front.View{Party: []front.Member{{Name: "A", HP: 5, MaxHP: 5, Selected: true}}, Sheet: &front.Sheet{},
		SheetPage: &front.SheetPage{Weapon: "LONGSWORD", Armour: "CHAIN MAIL", Encumbrance: "NORMAL"}}
	(&Presenter{bars: gpu.NewImageFromImage(barArt())}).drawSheet(dst, note, v)
	for _, w := range []string{"LONGSWORD", "CHAIN MAIL", "NORMAL", "Load"} {
		if !strings.Contains(drawn.String(), w) {
			t.Errorf("%q not drawn", w)
		}
	}
}
