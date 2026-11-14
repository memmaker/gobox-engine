package ui

import (
	"testing"

	"gobox/game"
	"gobox/gpu"
)

// TestSheetLeftValuesTouchBorder: the left pane's values (HP, Status)
// end at the pane's border, touching it, never under it.
func TestSheetLeftValuesTouchBorder(t *testing.T) {
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
	touch := false
	for y := px.Rect.Min.Y; y < px.Rect.Max.Y; y++ {
		touch = touch || px.RGBAAt(edge-1, y) == green
		if px.RGBAAt(edge, y) == green {
			t.Fatalf("a value at x %d, under the border", edge)
		}
	}
	if !touch {
		t.Fatalf("no value touches the border at %d", edge)
	}
}
