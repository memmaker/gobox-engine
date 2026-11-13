package present

import (
	"image"
	"strings"
	"testing"

	"gobox/front"
	"gobox/gpu"
	"gobox/ui"
)

// TestCreditsCellSize (CLAUDE.md, one glyph cell size): the title's credits
// page, with the original's boxes (POR) and without (one frame), fills the
// window at K at several window sizes; its lines and the remaster's stay
// inside the page, in order, and on a tall page the sections spread down the
// box (por-go b77c46ea) with the remaster's lines under them.
func TestCreditsCellSize(t *testing.T) {
	defer func(l []Credit, b []image.Rectangle, r [2]string, g []int) {
		credits, creditBoxes, remasterCredit, creditGaps, drawnAt = l, b, r, g, nil
	}(credits, creditBoxes, remasterCredit, creditGaps)
	credits = []Credit{ // POR's (ovr02:0206), abridged: the sections and their blank rows
		{X: 5, Y: 1, Colour: 0xa, Text: "scenario created by:"},
		{X: 2, Y: 3, Colour: 0xa, Text: "david cook"},
		{X: 1, Y: 5, Colour: 0xa, Text: "game created by:"},
		{X: 0xe, Y: 6, Colour: 0xb, Text: "version 1.3"},
		{X: 2, Y: 7, Colour: 0xe, Text: "programming:"},
		{X: 2, Y: 0xb, Colour: 0xb, Text: "ted greer"},
		{X: 2, Y: 0xd, Colour: 0xe, Text: "graphic arts:"},
		{X: 2, Y: 0x13, Colour: 0xe, Text: "project manager"},
		{X: 2, Y: 0x17, Colour: 0xb, Text: "james kucera"},
	}
	remasterCredit, creditGaps = [2]string{"Opinionated Remaster", "directed by"}, []int{4, 5, 6}
	note := Note{Glyph: func(byte) [8]byte { return [8]byte{255} }, Font: func(byte, bool) [8]byte { return [8]byte{255} }}
	for _, boxes := range [][]image.Rectangle{{{Min: image.Pt(1, 1), Max: image.Pt(38, 3)}, {Min: image.Pt(1, 5), Max: image.Pt(38, 23)}}, nil} {
		creditBoxes = boxes
		for _, sz := range [][2]int{{800, 500}, {1280, 800}, {1920, 1200}, {2560, 1440}, {3840, 2400}} {
			dst := gpu.NewImage(sz[0], sz[1])
			K := float64(ui.K(dst.Bounds()))
			pages := 0
			drawnAt = func(what drawnKind, on *gpu.Image, k float64) {
				if what == drawnPage && on == dst {
					pages++
					if k != K {
						t.Errorf("%v boxes %v: credits page at %v, K %v", sz, boxes != nil, k, K)
					}
				}
			}
			v := front.View{Credits: true, Cells: image.Pt(40, 25)} // the game's page left at 40x25: the presenter lays it out to the window
			(&Presenter{}).Draw(dst, note, v)
			if pages != 1 {
				t.Errorf("%v boxes %v: %d credits pages drawn", sz, boxes != nil, pages)
			}
			dst.Deallocate()

			v.Cells.X, v.Cells.Y = sz[0]/(8*int(K)), sz[1]/(8*int(K))
			c := newCellPage(v)
			c.texts(v)
			var rows []string
			for _, r := range c.cells {
				var b strings.Builder
				for _, x := range r {
					if x.set {
						b.WriteByte(x.ch)
					}
				}
				rows = append(rows, b.String())
			}
			at := func(s string) int {
				for y, r := range rows {
					if strings.Contains(r, s) {
						return y
					}
				}
				return -1
			}
			prev := 0
			for _, l := range credits {
				y := at(l.Text)
				if y < prev || y > c.rows-3 {
					t.Errorf("%v boxes %v: %q at row %d of %d (after %d)", sz, boxes != nil, l.Text, y, c.rows, prev)
				}
				prev = y
			}
			if r := at(remasterCredit[0]); c.rows >= 30 && (r <= prev || r > c.rows-3) {
				t.Errorf("%v boxes %v: remaster at %d, last line %d, %d rows", sz, boxes != nil, r, prev, c.rows)
			}
			if c.rows >= 37 && c.rows-at("james kucera") > 7 {
				t.Errorf("%v boxes %v: the credits end at row %d of %d, not spread", sz, boxes != nil, prev, c.rows)
			}
		}
	}
}
