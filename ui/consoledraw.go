package ui

import (
	"image"
	"image/color"

	"gobox/asset"
)

// Draw renders the set character cells at 8x8 pixels a cell into dst (cell
// (0,0) at its top left) in font bits (nil: the game's): Colour on Bg, both
// looked up in the game's palette. Symbol cells use symbol(id), a game
// tile; without it they are skipped.
func (c *TextConsole) Draw(dst *image.RGBA, bits []byte, pal func(byte) color.RGBA, symbol func(id int) ([8]byte, bool)) {
	for y, row := range c.Cells {
		for x, cl := range row {
			if !cl.Set {
				continue
			}
			var g [8]byte
			if cl.Symbol {
				if symbol == nil {
					continue
				}
				g, _ = symbol(cl.Sym)
				bg := pal(0)
				fg := pal(cl.Colour)
				asset.DrawGlyph(dst, g, x*8, y*8, &fg, &bg)
				continue
			}
			g, _ = asset.Glyph(bits, cl.C, cl.Upcase)
			fg, bg := pal(cl.Colour), pal(cl.Bg)
			asset.DrawGlyph(dst, g, x*8, y*8, &fg, &bg)
		}
	}
}
