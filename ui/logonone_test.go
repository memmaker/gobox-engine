package ui

import (
	"image"
	"testing"

	"gobox/asset"
)

// TestAmigaRef: an Amiga block named by file and block (ColourBase 32, its
// own palette) decodes opaque in that palette (artRGBA left every Amiga
// pixel out: c < asset.Clear).
func TestAmigaRef(t *testing.T) {
	defer func(a *asset.Art, am bool) { daxArt.art, daxArt.amiga = a, am; clear(daxPics) }(daxArt.art, daxArt.amiga)
	daxArt.art, daxArt.amiga = asset.OpenArt("../data/cotab/art"), true
	clear(daxPics)
	b := artBlock("TITLE", 1)
	if b == nil || b.ColourBase < asset.AmigaBase || len(b.Palette) == 0 {
		t.Skip("no imported cotab Amiga TITLE 1")
	}
	m, _ := daxImage("TITLE", 1, 0).(*image.RGBA)
	if m == nil {
		t.Fatal("no image")
	}
	for i, c := range b.Frames[0].Image.Data {
		k := int(c-uint8(b.ColourBase)) * 3
		if got := m.RGBAAt(i%m.Rect.Dx(), i/m.Rect.Dx()); got.A != 255 || [3]uint8{got.R, got.G, got.B} != [3]uint8(b.Palette[k:k+3]) {
			t.Fatalf("pixel %d (index %d): %v, want palette %v", i, c, got, b.Palette[k:k+3])
		}
	}
}
