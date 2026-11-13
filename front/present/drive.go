// drive.go: the presenter driven from outside, as a game's own tests do:
// a decision on a shown screen, the screens it puts up, and what it lays out.

package present

import (
	"image"
	"image/color"

	"gobox/front"
)

// Decide takes a on the screen shown (as a key, click or pad press would).
func (p *Presenter) Decide(shown front.View, a front.Action) {
	p.shownView = shown
	p.sendA(a)
}

// Own is v with the presenter's own screens up.
func (p *Presenter) Own(v front.View) front.View { return p.own(v) }

// IconClick is the action a click on cell cx, cy of the icon editor takes.
func IconClick(ed *front.IconEditor, cx, cy int) (front.Action, bool) { return iconClick(ed, cx, cy) }

// MapImage draws m at cell pixels a square in the colours ink.
func MapImage(m *front.MapSquares, cell int, ink [16]color.RGBA, glyph func(byte, bool) [8]byte) *image.RGBA {
	return mapImage(m, cell, ink, glyph)
}
