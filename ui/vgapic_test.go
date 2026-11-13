package ui

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"gobox/asset/orig"
)

// A VGA DAX picture (PoD's TITLE) is drawn in its own palette, not decoded as EGA
// (skipped without pod-go's data/dos). PNG: $VGAPIC_PNG.
func TestDAXPictureVGA(t *testing.T) {
	l, err := orig.NewLoader(filepath.Join("..", "data", "pod", "sources", "dos"))
	if err != nil || !l.Exists("TITLE.DAX") {
		t.Skip("no pod data")
	}
	useArt(t, l.Dir)
	clear(daxPics)
	want := orig.ParseVGA(l.BlockShared("TITLE", 1), false, 0)
	m := daxImage("TITLE", 1, 0)
	if p := os.Getenv("VGAPIC_PNG"); p != "" && m != nil {
		f, _ := os.Create(p)
		png.Encode(f, m)
		f.Close()
	}
	if want == nil || m == nil || m.Bounds().Dx() != want.Width*8 || m.Bounds().Dy() != want.Height {
		t.Fatalf("TITLE 1: got %v, want %dx%d", m != nil, want.Width*8, want.Height)
	}
	for i, c := range want.Frame(0) {
		if m.At(i%(want.Width*8), i/(want.Width*8)) != want.Palette[c] {
			t.Fatalf("pixel %d: %v, want %v", i, m.At(i%(want.Width*8), i/(want.Width*8)), want.Palette[c])
		}
	}
}
