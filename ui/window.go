package ui

import (
	"image"
	"image/color"

	"gobox/gpu"
	"gobox/gpu/vector"
)

// Ribbon pieces: a border's cells by their place (Ribbon).
const (
	TopLeft = iota
	TopRight
	BottomLeft
	BottomRight
	Top
	Bottom
	Left
	Right
	Pieces                // the corners and sides; a game's fixed [Pieces] literals stay valid
	LeftTopCap = iota - 1 // a side's first and last cell, just inside the corners; nil: the side piece
	LeftBottomCap
	RightTopCap
	RightBottomCap
	Places // every place, caps included
)

// Ribbon is a border's 8x8 pieces by place (TopLeft ..); a nil piece is
// not drawn.
type Ribbon [Places]*gpu.Image

// Window is every window's frame: a cw by ch cell box at g on dst, bg
// inside (nil: the inside stays as drawn), then the ribbon border in tiles
// t (a nil piece: none), as DrawBorder does.
// It returns the inside, in dst's pixels.
func Window(dst *gpu.Image, t Ribbon, g gpu.GeoM, cw, ch int, bg color.Color) image.Rectangle {
	x0, y0 := g.Apply(8, 8)
	x1, y1 := g.Apply(float64(cw*8-8), float64(ch*8-8))
	if bg != nil {
		vector.DrawFilledRect(dst, float32(x0), float32(y0), float32(x1-x0), float32(y1-y0), bg)
	}
	cell := func(i, x, y int) {
		if t[i] == nil {
			return
		}
		op := &gpu.DrawImageOptions{}
		op.GeoM.Translate(float64(x*8), float64(y*8))
		op.GeoM.Concat(g)
		dst.DrawImage(t[i], op)
	}
	for x := 1; x < cw-1; x++ {
		cell(Top, x, 0)
		cell(Bottom, x, ch-1)
	}
	pick := func(cap, side int, end bool) int {
		if end && t[cap] != nil {
			return cap
		}
		return side
	}
	for y := 1; y < ch-1; y++ {
		l, r := Left, Right
		if y == 1 {
			l, r = pick(LeftTopCap, l, true), pick(RightTopCap, r, true)
		}
		if y == ch-2 {
			l, r = pick(LeftBottomCap, l, true), pick(RightBottomCap, r, true)
		}
		cell(l, 0, y)
		cell(r, cw-1, y)
	}
	cell(TopLeft, 0, 0)
	cell(TopRight, cw-1, 0)
	cell(BottomLeft, 0, ch-1)
	cell(BottomRight, cw-1, ch-1)
	return image.Rect(int(x0), int(y0), int(x1), int(y1))
}
