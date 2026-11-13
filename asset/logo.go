package asset

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

// LogoCut says how a menu's lettering is cut out of a title picture (EGA
// indexes): above Sub the Keep colours in the EGA palette, from Sub to
// Bottom the SubInk pixels in Gold with an Outline round them; everything
// else transparent, cut to the lettering.
type LogoCut struct {
	Sub, Bottom   int // the subtitle's first row; the row under it
	Keep          []uint8
	SubInk        uint8
	Gold, Outline color.RGBA
}

// CutLogo is the lettering of pix (w pixels wide) by lc. A cut that finds no
// lettering is the cut or the picture wrong: it is an error.
func CutLogo(pix []uint8, w int, lc LogoCut) (image.Image, error) {
	if w == 0 || lc.Bottom == 0 {
		return nil, fmt.Errorf("logo cut %+v on a %d-wide picture", lc, w)
	}
	keep := func(c uint8) bool {
		for _, k := range lc.Keep {
			if c == k {
				return true
			}
		}
		return false
	}
	out := image.NewRGBA(image.Rect(0, 0, w, min(len(pix)/w, lc.Bottom)))
	var r image.Rectangle // the lettering's extent
	for i, c := range pix {
		x, y := i%w, i/w
		c &= 15
		if y < lc.Sub && keep(c) {
			out.Set(x, y, EGARGBA(int(c)))
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
		return nil, fmt.Errorf("logo cut %+v finds no lettering in the %d-wide picture", lc, w)
	}
	cut := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	draw.Draw(cut, cut.Bounds(), out, r.Min, draw.Src)
	return cut, nil
}
