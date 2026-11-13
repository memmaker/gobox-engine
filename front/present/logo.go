package present

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"gobox/ega"
	"gobox/front"
)

// LogoCut says how the main menu's lettering is cut out of the title
// picture (Logo's Pix, EGA indexes): above Sub the Keep colours in
// the EGA palette, from Sub to Bottom the SubInk pixels in Gold with an
// Outline round them; everything else transparent, cut to the lettering.
type LogoCut struct {
	Sub, Bottom   int // the subtitle's first row; the row under it
	Keep          []uint8
	SubInk        uint8
	Gold, Outline color.RGBA
}

var logoCache struct {
	pix *uint8
	img image.Image
}

// none: the game cuts no lettering (Data.Logo nil, the zero LogoCut): its
// title backdrop carries its own (COTAB's TITLE 3).
func (l LogoCut) none() bool { return l.Bottom == 0 }

// logoImage is p's lettering by Data.Logo (nil: none), cut once per picture.
// A cut that finds no lettering is the game's data wrong: it panics.
func logoImage(p *front.Picture) image.Image {
	if p == nil || len(p.Pix) == 0 || p.PixW == 0 || logoCut.none() {
		return nil
	}
	if logoCache.pix == &p.Pix[0] {
		return logoCache.img
	}
	lc, w := logoCut, p.PixW
	keep := func(c uint8) bool {
		for _, k := range lc.Keep {
			if c == k {
				return true
			}
		}
		return false
	}
	out := image.NewRGBA(image.Rect(0, 0, w, min(len(p.Pix)/w, lc.Bottom)))
	var r image.Rectangle // the lettering's extent
	for i, c := range p.Pix {
		x, y := i%w, i/w
		c &= 15
		if y < lc.Sub && keep(c) {
			out.Set(x, y, ega.RGBA(int(c)))
			r = r.Union(image.Rect(x, y, x+1, y+1))
		} else if y >= lc.Sub && y < lc.Bottom && c == lc.SubInk {
			out.Set(x, y, lc.Gold)
			r = r.Union(image.Rect(x-1, y-1, x+2, y+2))
		}
	}
	for y := lc.Sub - 1; y <= lc.Bottom; y++ { // the outline round the gold
		for x := 0; x < w; x++ {
			if out.RGBAAt(x, y).A != 0 {
				continue
			}
			for _, d := range [...]image.Point{{-1, -1}, {0, -1}, {1, -1}, {-1, 0}, {1, 0}, {-1, 1}, {0, 1}, {1, 1}} {
				if out.RGBAAt(x+d.X, y+d.Y) == lc.Gold {
					out.Set(x, y, lc.Outline)
					break
				}
			}
		}
	}
	if r.Empty() {
		panic(fmt.Sprintf("Data.Logo %+v cuts no lettering from the %d-wide title picture", lc, w))
	}
	cut := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	draw.Draw(cut, cut.Bounds(), out, r.Min, draw.Src)
	logoCache.pix, logoCache.img = &p.Pix[0], cut
	return cut
}
