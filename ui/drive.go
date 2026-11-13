// drive.go: the presenter driven from outside, as a game's own tests do:
// a decision on a shown screen, the screens it puts up, and what it lays out.

package ui

import (
	"image"
	"image/color"
)

// IconClick is the action a click on cell cx, cy of the icon editor takes.
func IconClick(ed *FrontIconEditor, cx, cy int) (FrontAction, bool) { return iconClick(ed, cx, cy) }

// MapImage draws m at cell pixels a square in the colours ink.
func MapImage(m *FrontMapSquares, cell int, ink [16]color.RGBA, glyph func(byte, bool) [8]byte) *image.RGBA {
	return mapImage(m, cell, ink, glyph)
}
