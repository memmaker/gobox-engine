package asset

import (
	"bytes"
	"image/color"
	"testing"
)

// TestSheetRoundTrip: a written sheet reads back pixel for pixel, a masked
// load clears ColourBase+mask, and a broken frame line fails.
func TestSheetRoundTrip(t *testing.T) {
	still := NewImage(2, 2, 3)
	for i := range still.Data {
		still.Data[i] = byte(32 + i%20)
	}
	a1, a2 := NewImage(1, 1, 2), NewImage(1, 3, 1)
	a1.X, a2.Y, a2.Data[5] = 4, 7, 9
	blocks := []Block{
		{ID: 5, Kind: Still, ColourBase: 32, Palette: []byte{1, 2, 3}, Frames: []Frame{{Image: frameOf(still, 0)}, {Image: frameOf(still, 1)}}},
		{ID: 2, Kind: Anim, Frames: []Frame{{Delay: 3, Image: a1}, {Delay: 8, Image: a2}}},
	}
	pal := make(color.Palette, 256)
	for i := range pal {
		pal[i] = color.Gray{byte(i)}
	}
	dir := t.TempDir()
	if err := WriteSheet(dir, "X", blocks, pal); err != nil {
		t.Fatal(err)
	}
	got, err := ReadSheet(dir, "X")
	if err != nil {
		t.Fatal(err)
	}
	im := got[5].Image(NoMask)
	if !bytes.Equal(im.Data, still.Data) || im.Count != 2 || !bytes.Equal(got[5].Palette, []byte{1, 2, 3}) {
		t.Fatalf("still: %v", im)
	}
	m := got[5].Image(1)
	for i, p := range still.Data {
		if want := map[bool]byte{true: Clear, false: p}[p == 33]; m.Data[i] != want {
			t.Fatalf("masked pixel %d = %d, want %d", i, m.Data[i], want)
		}
	}
	an := got[2].Anim(NoMask)
	if len(an) != 2 || an[1].Delay != 8 || an[0].Image.X != 4 || an[1].Image.Y != 7 || an[1].Image.Data[5] != 9 || an[1].Image.Width != 3 {
		t.Fatalf("anim: %+v %+v", an[0], an[1])
	}
}

func frameOf(im *Image, i int) *Image {
	f := NewImage(1, im.Width, im.Height)
	copy(f.Data, im.Frame(i))
	return f
}

// TestAnimOwnColours: an Amiga animation block (cotab's PIC 10, the marked
// sword arm) plays in its own palette (Block.colours, as a still); it fell back to POR's
// asset.Amiga slots (tan skin, green and blue speckles).
func TestAnimOwnColours(t *testing.T) {
	rgb := []byte{0, 0, 0, 0xaa, 0x88, 0x00} // colour 1: the fixed palette's olive
	b := Block{Kind: Anim, ColourBase: AmigaBase, Palette: rgb, Frames: []Frame{{Image: NewImage(1, 1, 1)}}}
	want := color.RGBA{0xaa, 0x88, 0x00, 255}
	if got := b.Anim(NoMask)[0].Image.Palette; got == nil || got[AmigaBase+1] != want {
		t.Fatalf("animation: colour %d not the block's own (palette %v)", AmigaBase+1, got == nil)
	}
}
