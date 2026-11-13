// classicpic.go: the game's own pixels (Picture.Pix) in the page's ink.

package ui

import (
	"image"
	"image/color"

	"gobox/ega"
)

type pixKey struct {
	pix  string
	w    int
	mask int
	ink  [16]color.RGBA
	pal  *color.RGBA // the palette's first entry (a game shares one slice), nil for EGA
}

var pixPics = map[pixKey]image.Image{}

// egaInk is the EGA's own colours, for the pictures given in them.
var egaInk = func() (k [16]color.RGBA) {
	for i := range k {
		k[i] = ega.RGBA(i)
	}
	return k
}()

// pixImage is pc's pixels in ink (nil without pixels), cached.
func pixImage(pc Picture, ink [16]color.RGBA) image.Image {
	if pc.Pix == nil {
		return nil
	}
	if len(pixPics) > 4096 { // ponytail: whole cache dropped when big
		clear(pixPics)
	}
	k := pixKey{string(pc.Pix), pc.PixW, -1, ink, nil}
	if pc.Palette != nil {
		k.pal = &pc.Palette[0]
	}
	if pc.Mask != nil {
		k.mask = int(*pc.Mask)
	}
	if m, ok := pixPics[k]; ok {
		return m
	}
	var m *image.RGBA
	if pc.Palette != nil {
		m = PalettedRGBA(pc.Pix, pc.PixW, pc.Palette)
	} else {
		m = IndexedRGBA(pc.Pix, pc.PixW, k.mask, &ink)
	}
	pixPics[k] = m
	return m
}

// picImage is pc as an image: its pixels in ink, else the DAX picture its
// File, Block and Frame name (EGA colours), else nil.
func picImage(pc Picture, ink [16]color.RGBA) image.Image {
	if pc.Img != nil {
		return pc.Img
	}
	if ink == egaInk && pc.Palette == nil {
		if m := poolImage(pc); m != nil {
			return m
		}
	}
	if pc.Pix != nil {
		return pixImage(pc, ink)
	}
	if pc.File != "" {
		return daxMasked(pc.File, pc.Block, pc.Frame, refMask(pc))
	}
	return nil
}

// nativeSize is pc's size in the original's pixels (24x24 a square), whatever
// remastered picture stands in for it (zero: unknown).
func nativeSize(pc Picture) image.Point {
	switch {
	case pc.PixW > 0:
		return image.Pt(pc.PixW, len(pc.Pix)/pc.PixW)
	case pc.File != "":
		if m := daxMasked(pc.File, pc.Block, pc.Frame, refMask(pc)); m != nil {
			return m.Bounds().Size()
		}
	}
	return image.Point{}
}

// indexedPictures is ps with the pictures given as pixels converted in ink.
func indexedPictures(ps []Picture, ink [16]color.RGBA) []Picture {
	var out []Picture
	for i, pc := range ps {
		if pc.Pix == nil || pc.Img != nil {
			continue
		}
		if out == nil {
			out = append([]Picture(nil), ps...)
		}
		out[i].Img = pixImage(pc, ink)
	}
	if out == nil {
		return ps
	}
	return out
}

// textured is t, or where t is one flat colour (a Pool floor tile is a
// plain fill) the series' floor texture tinted to that colour: no combat
// square is ever a flat fill.
func textured(t image.Image) image.Image {
	if t == nil {
		return nil
	}
	k := texturedKey{t, series}
	if m, ok := texturedTiles[k]; ok { // the board is rebuilt every frame: each tile is looked at once
		return m
	}
	if len(texturedTiles) > 4096 { // ponytail: whole cache dropped when big, as pixPics
		clear(texturedTiles)
	}
	m := texturedOnce(t)
	texturedTiles[k] = m
	return m
}

type texturedKey struct {
	tile   image.Image
	series string
}

var texturedTiles = map[texturedKey]image.Image{}

func texturedOnce(t image.Image) image.Image {
	c, ok := flatColour(t)
	if !ok {
		return t
	}
	tex := wallImage("floor")
	if tex == nil {
		return t
	}
	if m, ok := flatTiles[c]; ok {
		return m
	}
	m := Tint(tex, c)
	flatTiles[c] = m
	return m
}

var flatTiles = map[color.RGBA]image.Image{}

// flatColour is t's one colour when every texel is that colour, opaque.
func flatColour(t image.Image) (color.RGBA, bool) {
	if t == nil {
		return color.RGBA{}, false
	}
	b := t.Bounds()
	first := color.RGBAModel.Convert(t.At(b.Min.X, b.Min.Y)).(color.RGBA)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if color.RGBAModel.Convert(t.At(x, y)).(color.RGBA) != first {
				return first, false
			}
		}
	}
	return first, first.A == 255
}
