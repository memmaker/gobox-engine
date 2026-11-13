package ui

import (
	"image"
	"image/color"
)

// mapImage draws m at cell pixels a square in the colours ink (ui.MapImage:
// the engine's squares copied over, their Seen and wall values map 1:1).
func mapImage(m *FrontMapSquares, cell int, ink [16]color.RGBA, glyph func(byte, bool) [8]byte) *image.RGBA {
	u := &MapSquares{Party: m.Party, Ahead: m.Ahead}
	for y, row := range m.Squares {
		for x, q := range row {
			s := MapSquare{Seen: Seen(q.Seen), Letter: q.Letter}
			if q.Letter != 0 {
				s.Ink = labelInks[q.Kind]
			}
			for d, w := range q.Walls {
				s.Walls[d] = Wall(w)
			}
			u.Squares[y][x] = s
		}
	}
	return UIMapImage(u, cell, ink, doorRGBA, glyph)
}
