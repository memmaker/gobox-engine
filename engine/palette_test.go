package engine

import (
	"image/color"
	"testing"

	"gobox/asset"
)

// TestArtKeepsPalette: where an Amiga image becomes View pixels (combat icons
// and tiles, the border's pieces; the icon editor's poses and portrait heads
// the same way) it carries the image's own colours (Image.Palette, set by
// loadPic from the block): without them the presenter draws cotab/sotsb/pod's
// Amiga art in POR's asset.Amiga slots.
func TestArtKeepsPalette(t *testing.T) {
	pal := make([]color.RGBA, 256)
	pal[asset.AmigaBase] = color.RGBA{1, 2, 3, 255}
	im := &asset.Image{Height: 2, Width: 1, Count: 1, BPP: 16, Data: make([]byte, 16), Palette: pal}
	same := func(what string, got []color.RGBA) {
		t.Helper()
		if len(got) != len(pal) || got[asset.AmigaBase] != pal[asset.AmigaBase] {
			t.Errorf("%s: palette lost", what)
		}
	}
	var e Engine
	same("picPix", e.picPix(im, 0, false, nil).Palette)
	same("tilePix", e.tilePix(im, 0, nil).Palette)
	sym := *im
	sym.Count, sym.Data = 0x17, make([]byte, 16*0x17)
	e.sym[4] = &sym
	for _, p := range e.borderTiles() {
		same("borderTiles", p.Palette)
	}
}
