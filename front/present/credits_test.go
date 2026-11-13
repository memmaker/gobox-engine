package present

import (
	"encoding/json"
	"image"
	"os"
	"strings"
	"testing"

	"gobox/front"
	"gobox/gpu"
	"gobox/importer"
	"gobox/ui"
)

var creditSizes = [][2]int{{800, 500}, {1280, 800}, {1920, 1200}, {2560, 1440}, {3840, 2400}}

// creditsGames are the four games' imported manifests (data/<game>).
var creditsGames = []string{"por", "cotab", "sotsb", "pod"}

// remasterCreditText is the line Felix asked for on every credits screen.
const remasterCreditText = "Opinionated Remaster directed by Felix Ruzzoli using Claude"

// checkCreditsPage lays the credits page out at every size and checks it: at
// K, inside the window, the credits above the remaster credit, the remaster credit exactly once
// below them, alone on its rows (one row or two), above the frame's edge.
func checkCreditsPage(t *testing.T, name string, lines []Credit, boxes []image.Rectangle) {
	defer func(l []Credit, b []image.Rectangle, g []int) {
		credits, creditBoxes, creditGaps, drawnAt = l, b, g, nil
	}(credits, creditBoxes, creditGaps)
	credits, creditBoxes = lines, boxes
	note := Note{Glyph: func(byte) [8]byte { return [8]byte{255} }, Font: func(byte, bool) [8]byte { return [8]byte{255} }}
	for _, sz := range creditSizes {
		dst := gpu.NewImage(sz[0], sz[1])
		K := float64(ui.K(dst.Bounds()))
		pages := 0
		drawnAt = func(what drawnKind, on *gpu.Image, k float64) {
			if what == drawnPage && on == dst {
				pages++
				if k != K {
					t.Errorf("%s %v: credits page at %v, K %v", name, sz, k, K)
				}
			}
		}
		v := front.View{Credits: true, Cells: image.Pt(40, 25)} // the game's page left at 40x25: the presenter lays it out to the window
		(&Presenter{}).Draw(dst, note, v)
		if pages != 1 {
			t.Errorf("%s %v: %d credits pages drawn", name, sz, pages)
		}
		dst.Deallocate()

		v.Cells.X, v.Cells.Y = sz[0]/(8*int(K)), sz[1]/(8*int(K))
		c := newCellPage(v)
		c.texts(v, false)
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
		first := -1 // the remaster credit's first row
		for y, r := range rows {
			if strings.Contains(r, "Opinionated") {
				if first >= 0 {
					t.Errorf("%s %v: remaster credit twice (rows %d, %d)", name, sz, first, y)
				}
				first = y
			}
		}
		want := remasterRows(c.cols)
		if first < 0 || first+len(want) > c.rows-2 {
			t.Fatalf("%s %v: remaster credit at row %d of %d", name, sz, first, c.rows)
		}
		if strings.Join(want, " ") != remasterCreditText {
			t.Fatalf("remaster credit %q", want)
		}
		for i, w := range want {
			if rows[first+i] != w { // alone on its row: no credit line overlaps it
				t.Errorf("%s %v: row %d is %q, want %q", name, sz, first+i, rows[first+i], w)
			}
		}
		for _, l := range lines {
			y := -1
			for r, s := range rows {
				if strings.Contains(s, l.Text) && r < first {
					y = r
					break
				}
			}
			if y < 0 {
				t.Errorf("%s %v: %q not above the remaster credit (row %d)", name, sz, l.Text, first)
			}
		}
	}
}

// TestCreditsCellSize (CLAUDE.md, one glyph cell size): the title's credits
// page, with the original's boxes (POR) and without (one frame), fills the
// window at K at several window sizes; its lines and the remaster's stay
// inside the page, in order, and on a tall page the sections spread down the
// box (por-go b77c46ea) with the remaster's line under them, once. The four
// games' imported pages (data/<game>) too.
func TestCreditsCellSize(t *testing.T) {
	defer func(g []int) { creditGaps = g }(creditGaps)
	creditGaps = []int{4, 5, 6}
	abridged := []Credit{ // POR's (ovr02:0206), abridged: the sections and their blank rows
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
	porBoxes := []image.Rectangle{{Min: image.Pt(1, 1), Max: image.Pt(38, 3)}, {Min: image.Pt(1, 5), Max: image.Pt(38, 23)}}
	checkCreditsPage(t, "abridged boxed", abridged, porBoxes)
	checkCreditsPage(t, "abridged edges", abridged, nil)
	for _, id := range creditsGames {
		b, err := os.ReadFile("../../data/" + id + "/manifest.json")
		if err != nil {
			t.Logf("%s: no imported manifest: %v", id, err)
			continue
		}
		var m importer.Manifest
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		if m.Start == nil || len(m.Start.Credits) == 0 {
			continue // a credits picture (pod): TestCreditsPicture
		}
		var lines []Credit
		for _, c := range m.Start.Credits {
			lines = append(lines, Credit{X: c.X, Y: c.Y, Colour: c.Colour, Text: c.Text, Page: c.Page})
		}
		var boxes []image.Rectangle
		for _, c := range m.Start.Boxes {
			boxes = append(boxes, image.Rect(c.X1, c.Y1, c.X2, c.Y2))
		}
		checkCreditsPage(t, id, lines, boxes)
	}
}

// TestCreditsPicture: a credits picture (PoD TITLE 3, SotSB's Amiga card)
// leaves a K-sized strip of the grid's width at the window's bottom for the
// remaster credit: inside the window, 8K per row, the picture above it
// fitted into what remains so it covers none of the picture.
func TestCreditsPicture(t *testing.T) {
	defer func() { drawnAt = nil }()
	note := Note{Glyph: func(byte) [8]byte { return [8]byte{255} }, Font: func(byte, bool) [8]byte { return [8]byte{255} }}
	for _, sz := range creditSizes {
		dst := gpu.NewImage(sz[0], sz[1])
		K := ui.K(dst.Bounds())
		strip := creditStrip(dst.Bounds())
		cols := sz[0] / (8 * K)
		if !strip.In(dst.Bounds()) || strip.Dy() != len(remasterRows(cols))*8*K || strip.Dx() != cols*8*K || strip.Max.Y != sz[1] {
			t.Errorf("%v: strip %v, K %d", sz, strip, K)
		}
		strips := 0
		drawnAt = func(what drawnKind, on *gpu.Image, k float64) {
			if what == drawnPage && on == dst {
				strips++
				if k != float64(K) {
					t.Errorf("%v: strip at %v, K %d", sz, k, K)
				}
			}
		}
		img := image.NewRGBA(image.Rect(0, 0, 320, 200))
		for i := range img.Pix {
			img.Pix[i] = 255
		}
		v := front.View{CreditsPic: true, Intro: &front.Intro{}, Cells: image.Pt(40, 25),
			Pictures: []front.Picture{{Img: img, File: "TITLE", At: img.Rect}}}
		(&Presenter{}).drawGeneric(dst, v, note)
		if strips != 1 {
			t.Errorf("%v: %d strips drawn", sz, strips)
		}
		dst.Deallocate()
	}
}
