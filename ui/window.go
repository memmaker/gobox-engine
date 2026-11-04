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
	Pieces
)

// Ribbon is a border's 8x8 pieces by place (TopLeft ..); a nil piece is
// not drawn.
type Ribbon [Pieces]*gpu.Image

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
	for y := 1; y < ch-1; y++ {
		cell(Left, 0, y)
		cell(Right, cw-1, y)
	}
	cell(TopLeft, 0, 0)
	cell(TopRight, cw-1, 0)
	cell(BottomLeft, 0, ch-1)
	cell(BottomRight, cw-1, ch-1)
	return image.Rect(int(x0), int(y0), int(x1), int(y1))
}
