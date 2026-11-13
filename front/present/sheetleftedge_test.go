package present

import (
	"testing"

	"gobox/game"
	"gobox/gpu"
)

// TestSheetLeftValuesClearBorder: the left pane's values (HP, Status)
// end a cell short of the pane's border, never touching it.
func TestSheetLeftValuesClearBorder(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	g := func(c byte) [8]byte {
		if c == '#' {
			return [8]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
		}
		return [8]byte{}
	}
	note := Note{Glyph: g, Font: func(c byte, _ bool) [8]byte { return g(c) }}
	p := sheetOn(&game.SheetPage{Status: "####", THAC0: 9})
	p.drawSheet(dst, note)
	px, edge := p.sheet.px, (leftCells-1)*8 // the border's inner edge
	for y := px.Rect.Min.Y; y < px.Rect.Max.Y; y++ {
		for x := edge - 8; x < edge; x++ {
			if c := px.RGBAAt(x, y); c == green {
				t.Fatalf("a value at x %d, within a cell of the border at %d", x, edge)
			}
		}
	}
}
