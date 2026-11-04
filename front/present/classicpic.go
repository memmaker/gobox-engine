// classicpic.go: the game's own pixels (Picture.Pix) in the page's ink.

package present

import (
	"image"
	"image/color"

	"gobox/front"
)

type pixKey struct {
	pix  string
	w    int
	mask int
}

var (
	pixPics = map[pixKey]image.Image{}
	pixInk  [16]color.RGBA
)

// indexedPictures is ps with the pictures given as pixels converted in ink.
func indexedPictures(ps []front.Picture, ink [16]color.RGBA) []front.Picture {
	if ink != pixInk || len(pixPics) > 512 { // ponytail: whole cache dropped on a palette change or when big
		clear(pixPics)
		pixInk = ink
	}
	var out []front.Picture
	for i, pc := range ps {
		if pc.Pix == nil || pc.Img != nil {
			continue
		}
		if out == nil {
			out = append([]front.Picture(nil), ps...)
		}
		k := pixKey{string(pc.Pix), pc.PixW, -1}
		if pc.Mask != nil {
			k.mask = int(*pc.Mask)
		}
		if m, ok := pixPics[k]; ok {
			out[i].Img = m
			continue
		}
		out[i] = pc.Converted(&ink)
		pixPics[k] = out[i].Img
	}
	if out == nil {
		return ps
	}
	return out
}
