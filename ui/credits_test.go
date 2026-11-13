package ui

import (
	"bytes"
	"encoding/json"
	"gobox/adventure"
	"gobox/game"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobox/gpu"
	"gobox/importer"
)

var creditSizes = [][2]int{{800, 500}, {1280, 800}, {1920, 1200}, {2560, 1440}, {3840, 2400}}

// creditsGames are the four games' imported manifests (data/<game>).
var creditsGames = []string{"por", "cotab", "sotsb", "pod"}

// remasterCreditText is the line Felix asked for on every credits screen.
const remasterCreditText = "Opinionated Remaster directed by Felix Ruzzoli using Claude"

// checkCreditsPage lays the credits page out at every size and checks it: at
// K, inside the window, the credits above the remaster credit, the remaster credit exactly once
// below them, alone on its rows (one row or two), above the frame's edge.
func checkCreditsPage(t *testing.T, name string, lines []adventure.Credit, boxes []adventure.Cells) {
	defer func(s *adventure.Sequence, g []int) {
		startSeq, creditGaps, drawnAt = s, g, nil
	}(startSeq, creditGaps)
	startSeq = &adventure.Sequence{Credits: lines, Boxes: boxes}
	note := Note{Glyph: func(byte) [8]byte { return [8]byte{255} }, Font: func(byte, bool) [8]byte { return [8]byte{255} }}
	for _, sz := range creditSizes {
		dst := gpu.NewImage(sz[0], sz[1])
		K := float64(K(dst.Bounds()))
		pages := 0
		drawnAt = func(what drawnKind, on *gpu.Image, k float64) {
			if what == drawnPage && on == dst {
				pages++
				if k != K {
					t.Errorf("%s %v: credits page at %v, K %v", name, sz, k, K)
				}
			}
		}
		p := creditsOn(image.Rect(0, 0, 320, 200)) // the presenter lays the page out to the window it draws in
		p.Draw(dst, note)
		if pages != 1 {
			t.Errorf("%s %v: %d credits pages drawn", name, sz, pages)
		}
		dst.Deallocate()

		if p.cells() != image.Pt(sz[0]/(8*int(K)), sz[1]/(8*int(K))) {
			t.Errorf("%s %v: page %v cells", name, sz, p.cells())
		}
		c := newCellPage(p.cells(), p.ink(), nil)
		c.texts(nil, p.flags(), false)
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
	abridged := []adventure.Credit{ // POR's (ovr02:0206), abridged: the sections and their blank rows
		{X: 5, Y: 1, Colour: game.InkLightGreen, Text: "scenario created by:"},
		{X: 2, Y: 3, Colour: game.InkLightGreen, Text: "david cook"},
		{X: 1, Y: 5, Colour: game.InkLightGreen, Text: "game created by:"},
		{X: 0xe, Y: 6, Colour: game.InkLightCyan, Text: "version 1.3"},
		{X: 2, Y: 7, Colour: game.InkYellow, Text: "programming:"},
		{X: 2, Y: 0xb, Colour: game.InkLightCyan, Text: "ted greer"},
		{X: 2, Y: 0xd, Colour: game.InkYellow, Text: "graphic arts:"},
		{X: 2, Y: 0x13, Colour: game.InkYellow, Text: "project manager"},
		{X: 2, Y: 0x17, Colour: game.InkLightCyan, Text: "james kucera"},
	}
	porBoxes := []adventure.Cells{{X1: 1, Y1: 1, X2: 38, Y2: 3}, {X1: 1, Y1: 5, X2: 38, Y2: 23}}
	checkCreditsPage(t, "abridged boxed", abridged, porBoxes)
	checkCreditsPage(t, "abridged edges", abridged, nil)
	for _, id := range creditsGames {
		b, err := os.ReadFile("../data/" + id + "/manifest.json")
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
		checkCreditsPage(t, id, m.Start.Credits, m.Start.Boxes)
	}
}

// TestCreditsPicture: a credits picture (PoD TITLE 3, SotSB's Amiga card)
// leaves a K-sized strip of the grid's width at the window's bottom for the
// remaster credit: inside the window, 8K per row, the picture above it
// fitted into what remains so it covers none of the picture.
func TestCreditsPicture(t *testing.T) {
	defer func() { drawnAt = nil }()
	dir := t.TempDir() // the credits picture: a white PNG in the adventure's art/
	if err := os.MkdirAll(filepath.Join(dir, adventure.ArtDir), 0o755); err != nil {
		t.Fatal(err)
	}
	white := image.NewRGBA(image.Rect(0, 0, 320, 200))
	draw.Draw(white, white.Bounds(), image.White, image.Point{}, draw.Src)
	var b bytes.Buffer
	if err := png.Encode(&b, white); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, adventure.ArtDir, "CREDITS.PNG"), b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	defer func(d string) { advDir = d }(advDir)
	advDir = dir
	withScripts(t, &adventure.Sequence{Steps: []adventure.SequenceStep{{Kind: adventure.StepPicture, Credits: true,
		Layers: []adventure.Layer{{Art: adventure.Art{File: "CREDITS.PNG", Kind: adventure.ArtPNG}, At: adventure.Cells{X2: 39, Y2: 24}}}}}}, nil)
	note := Note{Glyph: func(byte) [8]byte { return [8]byte{255} }, Font: func(byte, bool) [8]byte { return [8]byte{255} }}
	for _, sz := range creditSizes {
		dst := gpu.NewImage(sz[0], sz[1])
		K := K(dst.Bounds())
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
		p, _ := presenterOn(game.Scene{Kind: game.SceneTitle, Full: true, Intro: &game.Intro{Pictures: []int{1}}})
		p.drawGeneric(dst, note)
		if strips != 1 {
			t.Errorf("%v: %d strips drawn", sz, strips)
		}
		dst.Deallocate()
	}
}
