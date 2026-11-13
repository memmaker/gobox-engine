package ui

import (
	"image"
	"image/color"

	"gobox/gpu"
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
		gpu.DrawFilledRect(dst, float32(x0), float32(y0), float32(x1-x0), float32(y1-y0), bg)
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

// Edges are a border's cells one by one per edge, corners included (the
// original's cracked borders); a nil edge is drawn by the Ribbon.
type Edges struct{ Top, Bottom, Left, Right, Divider []*gpu.Image } // Divider: a vertical rule between two boxes

// EdgeCell is the entry of an m-entry edge sequence for cell i of n: the
// ends fixed, the middle entries repeated. An edge no longer than its
// table takes the table's head and tail (both ends' own cells kept).
func EdgeCell(i, n, m int) int {
	switch {
	case n <= m && i < n-1-i:
		return i
	case n <= m:
		return m - 1 - (n - 1 - i)
	case m <= 2 || i == 0:
		return min(i, m-1)
	case i == n-1:
		return m - 1
	}
	return (i-1)%(m-2) + 1
}

// WindowEdges draws e over a cw by ch cell window's border at g (after
// Window): the verticals, then the horizontals over the corners.
func WindowEdges(dst *gpu.Image, e Edges, g gpu.GeoM, cw, ch int) {
	run := func(seq []*gpu.Image, n int, at func(i int) (int, int)) {
		for i := 0; i < n && len(seq) > 0; i++ {
			t := seq[EdgeCell(i, n, len(seq))]
			if t == nil {
				continue
			}
			x, y := at(i)
			op := &gpu.DrawImageOptions{}
			op.GeoM.Translate(float64(x*8), float64(y*8))
			op.GeoM.Concat(g)
			dst.DrawImage(t, op)
		}
	}
	run(e.Left, ch, func(i int) (int, int) { return 0, i })
	run(e.Right, ch, func(i int) (int, int) { return cw - 1, i })
	run(e.Top, cw, func(i int) (int, int) { return i, 0 })
	run(e.Bottom, cw, func(i int) (int, int) { return i, ch - 1 })
}

// WindowDivider draws e.Divider down column 0 of a ch-cell-high box at g:
// the rule two side-by-side windows share (nil: as drawn).
func WindowDivider(dst *gpu.Image, e Edges, g gpu.GeoM, ch int) {
	WindowEdges(dst, Edges{Left: e.Divider}, g, 1, ch)
}
