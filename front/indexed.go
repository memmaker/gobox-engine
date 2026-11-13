package front

import (
	"image"
	"image/color"

	"gobox/ega"
	"gobox/game"
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

// PalettedRGBA is pix, w wide, as indices into pal (alpha 0 is transparent).
func PalettedRGBA(pix []uint8, w int, pal []color.RGBA) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, w, len(pix)/w))
	for i, p := range pix {
		if int(p) < len(pal) {
			out.SetRGBA(i%w, i/w, pal[p])
		}
	}
	return out
}

// PixelsOf is game picture p's pixels as a Picture (Pix, Palette, Mask):
// the one conversion of a game.Picture's Pixels.
func PixelsOf(p game.Picture) Picture {
	f := Picture{Pix: p.Pixels, PixW: p.W, Palette: p.Palette}
	if p.Clear >= 0 {
		m := uint8(p.Clear)
		f.Mask = &m
	}
	return f
}

// Converted is p with Img made from its Pix (p itself without Pix).
func (p Picture) Converted(ink *[16]color.RGBA) Picture {
	if p.Pix == nil || p.Img != nil {
		return p
	}
	if p.Palette != nil {
		p.Img = PalettedRGBA(p.Pix, p.PixW, p.Palette)
		return p
	}
	mask := -1
	if p.Mask != nil {
		mask = int(*p.Mask)
	}
	p.Img = IndexedRGBA(p.Pix, p.PixW, mask, ink)
	return p
}
