package present

import (
	"gobox/front"
	"image"
	"image/color"

	"gobox/ui"
)

// mapImage draws m at cell pixels a square in the colours ink (ui.MapImage:
// the engine's squares copied over, their Seen and wall values map 1:1).
func mapImage(m *front.MapSquares, cell int, ink [16]color.RGBA, glyph func(byte, bool) [8]byte) *image.RGBA {
	u := &ui.MapSquares{Party: m.Party, Ahead: m.Ahead}
	for y, row := range m.Squares {
		for x, q := range row {
			s := ui.MapSquare{Seen: ui.Seen(q.Seen), Ink: q.Ink, Letter: q.Letter}
			for d, w := range q.Walls {
				s.Walls[d] = ui.Wall(w)
			}
			u.Squares[y][x] = s
		}
	}
	return ui.MapImage(u, cell, ink, doorRGBA, glyph)
}
