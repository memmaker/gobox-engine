package combat

import (
	"image"
	"testing"

	"gobox/gpu"
)

// TestPicturelessEffectDrawn: an effect whose art did not resolve (COTAB's
// clouds, a bare missile) has no picture; Draw skips it rather than panic.
func TestPicturelessEffectDrawn(t *testing.T) {
	dst := gpu.NewImage(320, 200)
	defer dst.Deallocate()
	icon := image.NewRGBA(image.Rect(0, 0, 24, 24))
	icon.Pix[3] = 255
	b := &Board{Tiles: []image.Image{image.NewRGBA(image.Rect(0, 0, tilePx, tilePx))},
		Units:   []Unit{{ID: 1, X: 5, Y: 5, Size: 1, Party: true, Icon: icon, Name: "A", HP: 9, MaxHP: 9}, {ID: 2, X: 7, Y: 5, Size: 1, Icon: icon, Name: "GRENDEL", HP: 60, MaxHP: 60}},
		Effects: []Effect{{X: 6, Y: 5}, {X: 6, Y: 6, Pic: icon}}}
	b.Ground[5][6].Overlay = 1 // and a ground overlay drawn under them
	var p Presenter
	p.Draw(dst, b)
	p.Draw(dst, b)
	if p.figs[b.Tiles[0]] == nil {
		t.Error("the overlay was not drawn")
	}
}
