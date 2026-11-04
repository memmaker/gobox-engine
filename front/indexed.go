package front

import (
	"image"
	"image/color"

	"gobox/ega"
)

// IndexedRGBA is a Picture's Pix, w wide, as the original blits them:
// EGA 0..15 in ink, 16 ink[0], 17 up to ega.AmigaBase and mask (-1:
// none) left out, from ega.AmigaBase the Amiga colours.
func IndexedRGBA(pix []uint8, w, mask int, ink *[16]color.RGBA) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, w, len(pix)/w))
	for i, p := range pix {
		switch {
		case int(p) == mask, p > 16 && p < ega.AmigaBase:
		case p == 16:
			out.SetRGBA(i%w, i/w, ink[0])
		case p >= ega.AmigaBase:
			out.SetRGBA(i%w, i/w, ega.AmigaRGBA(p))
		default:
			out.SetRGBA(i%w, i/w, ink[p])
		}
	}
	return out
}

// Converted is p with Img made from its Pix (p itself without Pix).
func (p Picture) Converted(ink *[16]color.RGBA) Picture {
	if p.Pix == nil || p.Img != nil {
		return p
	}
	mask := -1
	if p.Mask != nil {
		mask = int(*p.Mask)
	}
	p.Img = IndexedRGBA(p.Pix, p.PixW, mask, ink)
	return p
}
