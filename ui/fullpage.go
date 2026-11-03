package ui

import "image"

// FullPageCells is the geometry of a screen of its own over the whole
// window (the bestiary, the inventory): the integer scale k of a canvas of
// at least minW x minH pixels, and the canvas size in 8-pixel cells.
func FullPageCells(b image.Rectangle, minW, minH int) (k, cw, ch int) {
	k = max(1, min(b.Dx()/minW, b.Dy()/minH))
	return k, max(minW, b.Dx()/k) / 8, max(minH, b.Dy()/k) / 8
}
