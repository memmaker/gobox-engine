package ui

import (
	"image"

	"gobox/gpu"
)

// barColors are every scroll bar's colours (the inventory's look).
func barColors() BarColors {
	return BarColors{Groove: dark, Thumb: gray, Light: white, Dark: dark}
}

// menuBar draws the scroll bar of a menu (drawMenu, at scale k) whose
// first rows lines (rs, its rows' rectangles) show n lines from *top, in
// its right blank column, and takes its clicks and drags: they move *top
// and keep *sel (nil: none) among the lines shown.
func (p *Presenter) menuBar(dst *gpu.Image, rs []image.Rectangle, n, rows int, top, sel *int, k float64) {
	if h, ok := MenuBar(rs, n, rows, top, sel, k); ok {
		h.Bar.DrawOn(dst, barColors())
		p.scrolls.Add(h)
	}
}
