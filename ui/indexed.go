package ui

import (
	"image"
	"image/color"

	"gobox/asset"
	"gobox/game"
)

// IndexedRGBA is a Picture's Pix, w wide, as the original blits them:
// EGA 0..15 in ink, 16 ink[0], 17 up to asset.AmigaBase and mask (-1:
// none) left out, from asset.AmigaBase the Amiga colours.
func IndexedRGBA(pix []uint8, w, mask int, ink *[16]color.RGBA) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, w, len(pix)/w))
	for i, p := range pix {
		switch {
		case int(p) == mask, p > 16 && p < asset.AmigaBase:
		case p == 16:
			out.SetRGBA(i%w, i/w, ink[0])
		case p >= asset.AmigaBase:
			out.SetRGBA(i%w, i/w, asset.AmigaRGBA(p))
		default:
			out.SetRGBA(i%w, i/w, ink[p])
		}
	}
	return out
}

// PalettedRGBA is pix, w wide, as indices into pal (alpha 0 is transparent), mask (-1: none) left out as IndexedRGBA
// leaves it: a VGA palette's own entry at the mask index is opaque.
func PalettedRGBA(pix []uint8, w, mask int, pal []color.RGBA) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, w, len(pix)/w))
	for i, p := range pix {
		if int(p) != mask && int(p) < len(pal) {
			out.SetRGBA(i%w, i/w, pal[p])
		}
	}
	return out
}

// PixelsOf is game picture p's pixels as a Picture (Pix, Palette, Mask):
// the one conversion of a game.Picture's Pixels.
func PixelsOf(p game.Picture) Picture {
	p = shownPic(p)
	f := Picture{Pix: p.Pixels, PixW: p.W, Palette: p.Palette}
	if p.Clear >= 0 {
		m := uint8(p.Clear)
		f.Mask = &m
	}
	return f
}

// maskOf is p's transparent index, -1: none.
func maskOf(p Picture) int {
	if p.Mask == nil {
		return -1
	}
	return int(*p.Mask)
}

// Converted is p with Img made from its Pix (p itself without Pix).
func (p Picture) Converted(ink *[16]color.RGBA) Picture {
	if p.Pix == nil || p.Img != nil {
		return p
	}
	mask := maskOf(p)
	if p.Palette != nil {
		p.Img = PalettedRGBA(p.Pix, p.PixW, mask, p.Palette)
		return p
	}
	p.Img = IndexedRGBA(p.Pix, p.PixW, mask, ink)
	return p
}
