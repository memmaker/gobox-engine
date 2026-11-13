package main

import (
	"testing"

	"gobox/asset"
)

// handMoves is the most pixels a frame of the opening's hand (PIC2 block
// 10) differs from the first by: the DOS frames' 57. The Amiga's middle
// frame, a smaller delta, wobbled (written alone, its clear pixels black).
const handMoves = 64

// TestCotabHandSteady: every frame of the hand, DOS and Amiga, is the first
// one's size and differs from it only by the hand's small movement.
func TestCotabHandSteady(t *testing.T) {
	for _, dir := range []string{"../../data/cotab/art/dos", "../../data/cotab/art/amiga"} {
		s, err := asset.ReadSheet(dir, "PIC2")
		if err != nil {
			t.Skip(err)
		}
		fr := s[10].Anim(-1)
		for k, f := range fr {
			a, b := fr[0].Image, f.Image
			if len(a.Data) != len(b.Data) {
				t.Errorf("%s: frame %d is %dx%d, the first %dx%d", dir, k, b.Width, b.Height, a.Width, a.Height)
				continue
			}
			n := 0
			for i := range a.Data {
				if a.Data[i] != b.Data[i] {
					n++
				}
			}
			if n > handMoves {
				t.Errorf("%s: frame %d differs from the first by %d pixels", dir, k, n)
			}
		}
	}
}
