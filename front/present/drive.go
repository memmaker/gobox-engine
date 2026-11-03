// drive.go: the presenter driven from outside, as a game's own tests do:
// a decision on a shown screen, the screens it puts up, and what it lays out.

package present

import (
	"fmt"
	"image"
	"image/color"
	"strings"

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

// ScreenDump is the page laid out for v, a line each: the text as runs
// (row, column, colours, upcase, text), the pictures (DAX file, block,
// frame, place; fills by colour; the border's 8x8 symbols left out) and the
// PNGs over the page (size, place).
func ScreenDump(v front.View) string {
	c := screenPage(v)
	var b strings.Builder
	hex := func(c color.RGBA) string { return fmt.Sprintf("%02x%02x%02x", c.R, c.G, c.B) }
	for y, l := range c.lines() {
		for _, r := range l {
			if strings.TrimSpace(r.Text) == "" && r.Bg == (color.RGBA{A: 255}) {
				continue
			}
			fmt.Fprintf(&b, "text %d,%d %s/%s %v %q\n", y, r.X, hex(r.Fg), hex(r.Bg), r.Upcase, r.Text)
		}
	}
	for _, p := range c.pics {
		switch {
		case p.Fill && p.Colour == color.RGBA{A: 255}: // the page's black
		case p.Fill:
			fmt.Fprintf(&b, "fill %v %s\n", p.At, hex(p.Colour))
		case p.At.Dx() == 8 && p.At.Dy() == 8:
		default:
			fmt.Fprintf(&b, "pic %s %d %d %v\n", p.File, p.Block, p.Frame, p.At)
		}
	}
	for _, o := range c.overs {
		fmt.Fprintf(&b, "png %v %v\n", o.Part().Size(), o.At)
	}
	return b.String()
}
