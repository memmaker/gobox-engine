package ui

import "image"

// PageW x PageH is the UI's smallest page in window pixels at K 1 (the
// character sheet's and the inventory's canvas): K is the largest whole
// factor it fits the window at.
const PageW, PageH = 400, 250

// K is the UI's one glyph cell size over 8 (CLAUDE.md): every glyph and
// ribbon tile is 8*K window pixels, K = max(1, min(W/PageW, H/PageH)). A
// larger window gives more cells, not bigger ones. It is the only place
// the UI's text scale is computed (TestOneK).
func K(b image.Rectangle) int { return max(1, min(b.Dx()/PageW, b.Dy()/PageH)) }

// FullPageCells is the geometry of a screen of its own over the whole
// window (the bestiary, the inventory, the sheet, the journal): K and the
// canvas, at least PageW x PageH pixels, in 8-pixel cells.
func FullPageCells(b image.Rectangle) (k, cw, ch int) {
	k = K(b)
	return k, max(PageW, b.Dx()/k) / 8, max(PageH, b.Dy()/k) / 8
}
