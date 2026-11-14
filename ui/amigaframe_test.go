package ui

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"path/filepath"
	"testing"

	"gobox/adventure"
	"gobox/asset"
	"gobox/game"
)

// TestAmigaFrameOverOwnPieces: POR's DOS border is the engine's own pieces
// (art "border/<n>"); under Amiga art the Amiga frame replaces them, every
// piece and edge cell inked, and under DOS art the own pieces stay.
func TestAmigaFrameOverOwnPieces(t *testing.T) {
	defer func(a *asset.Art, am bool, th Theme) {
		daxArt.art, daxArt.amiga = a, am
		clear(daxPics)
		clear(amigaScreens)
		applyTheme(th)
	}(daxArt.art, daxArt.amiga, theme)
	adv := filepath.Join("..", "data", "por")
	d, err := adventure.ReadGameData(adv)
	if err != nil {
		t.Skipf("por: %v", err)
	}
	if d.Border == nil || d.Border.Amiga == nil {
		t.Fatal("por: no Amiga frame in the game data")
	}
	daxArt.art = asset.OpenArt(filepath.Join(adv, "art"))
	if !daxArt.art.Has(asset.Amiga) {
		t.Skip("por: no imported Amiga art")
	}
	th := theme
	th.Border = Border{Amiga: d.Border.Amiga}
	applyTheme(th)
	const ownInk = 15
	own := game.Picture{W: frameCell, H: frameCell, Pixels: bytes.Repeat([]byte{ownInk}, frameCell*frameCell)}
	art := map[string]game.Picture{}
	for i := range LeftTopCap {
		art["border/"+string(rune('0'+i))] = own
	}
	for _, amiga := range []bool{false, true} {
		daxArt.amiga = amiga
		clear(daxPics)
		clear(amigaScreens)
		wb := borderOf(art, 0)
		if wb.own == amiga {
			t.Fatalf("amiga %v: own pieces %v", amiga, wb.own)
		}
		if !amiga {
			continue
		}
		for i, m := range wb.pieces[:LeftTopCap] {
			if !hasInk(m) {
				t.Errorf("piece %d empty", i)
			}
		}
		if len(wb.edges) != len(d.Border.Amiga.Edges) {
			t.Errorf("%d edges, want %d", len(wb.edges), len(d.Border.Amiga.Edges))
		}
		for k, cells := range wb.edges {
			for i, m := range cells {
				if !hasInk(m) {
					t.Errorf("edge %q cell %d empty", k, i)
				}
			}
		}
	}
}

// TestBorderSources: under each art (F12) and in every frame set, every
// border cell the front draws (the ribbon's pieces, each edge's cells) is
// a cell of that art's own frame with pixels in it: the Amiga frame's
// under Amiga art (no DOS symbol block read for it), the set's block.
// pod's sets differ from one another under both arts.
func TestBorderSources(t *testing.T) {
	defer func(a *asset.Art, am bool, th Theme) {
		daxArt.art, daxArt.amiga = a, am
		clear(daxPics)
		applyTheme(th)
	}(daxArt.art, daxArt.amiga, theme)
	for _, g := range []string{"cotab", "sotsb", "pod"} {
		adv := filepath.Join("..", "data", g)
		d, err := adventure.ReadGameData(adv)
		if err != nil {
			t.Skipf("%s: %v", g, err)
		}
		b := d.Border
		if b == nil || b.Amiga == nil {
			t.Fatalf("%s: no Amiga frame in the game data", g)
		}
		daxArt.art = asset.OpenArt(filepath.Join(adv, "art"))
		th := theme
		th.Border = Border{File: b.File, Block: b.Block, Pieces: b.Pieces, Caps: b.Caps, Mask: b.Mask, Edges: b.Edges, Sets: b.Sets, Amiga: b.Amiga}
		applyTheme(th)
		for _, amiga := range []bool{false, true} {
			daxArt.amiga = amiga
			clear(daxPics)
			clear(amigaScreens)
			sets := max(len(b.Sets), len(b.Amiga.Sets), 1)
			var tops [][]byte
			for set := range sets {
				wb := borderOf(nil, set)
				for i, m := range wb.pieces[:LeftTopCap] {
					if !hasInk(m) {
						t.Errorf("%s amiga %v set %d: piece %d empty", g, amiga, set, i)
					}
				}
				for k, cells := range wb.edges {
					for i, m := range cells {
						if !hasInk(m) {
							t.Errorf("%s amiga %v set %d: edge %q cell %d empty", g, amiga, set, k, i)
						}
					}
				}
				if amiga && len(wb.edges) != len(b.Amiga.Edges) {
					t.Errorf("%s set %d: %d Amiga edges, want %d", g, set, len(wb.edges), len(b.Amiga.Edges))
				}
				tops = append(tops, pixels(wb.pieces[0]))
			}
			for set := 1; set < len(tops); set++ {
				if bytes.Equal(tops[set], tops[0]) {
					t.Errorf("%s amiga %v: set %d's corner is set 0's", g, amiga, set)
				}
			}
		}
	}
}

// hasInk: m holds an opaque pixel.
func hasInk(m image.Image) bool {
	if m == nil {
		return false
	}
	for y := m.Bounds().Min.Y; y < m.Bounds().Max.Y; y++ {
		for x := m.Bounds().Min.X; x < m.Bounds().Max.X; x++ {
			if _, _, _, a := m.At(x, y).RGBA(); a != 0 {
				return true
			}
		}
	}
	return false
}

// pixels is m's RGBA bytes.
func pixels(m image.Image) []byte {
	if m == nil {
		return nil
	}
	r := image.NewRGBA(m.Bounds().Sub(m.Bounds().Min))
	draw.Draw(r, r.Bounds(), m, m.Bounds().Min, draw.Src)
	return r.Pix
}

// TestCotabMarkedBorder: cotab's DOS frame is black and white (greying
// marks nothing), so the selected card's border is its yellow twin
// (Border.MarkShift): every piece has yellow, the plain pieces none.
func TestCotabMarkedBorder(t *testing.T) {
	defer func(a *asset.Art, am bool, th Theme) {
		daxArt.art, daxArt.amiga = a, am
		clear(daxPics)
		applyTheme(th)
	}(daxArt.art, daxArt.amiga, theme)
	adv := filepath.Join("..", "data", "cotab")
	d, err := adventure.ReadGameData(adv)
	if err != nil {
		t.Skipf("cotab: %v", err)
	}
	b := d.Border
	daxArt.art, daxArt.amiga = asset.OpenArt(filepath.Join(adv, "art")), false
	clear(daxPics)
	th := theme
	th.Border = Border{File: b.File, Block: b.Block, Pieces: b.Pieces, Caps: b.Caps, Mask: b.Mask, Edges: b.Edges, Sets: b.Sets, MarkShift: b.MarkShift}
	applyTheme(th)
	wb := borderOf(nil, 0)
	yellow := func(m image.Image) bool {
		r := m.Bounds()
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				if c := color.RGBAModel.Convert(m.At(x, y)).(color.RGBA); c.A > 0 && c.R > 0xC0 && c.G > 0xC0 && c.B < 0x80 {
					return true
				}
			}
		}
		return false
	}
	for i := range LeftTopCap {
		if wb.marked[i] == nil || !yellow(wb.marked[i]) || yellow(wb.pieces[i]) {
			t.Errorf("piece %d: marked yellow %v, plain yellow %v", i, wb.marked[i] != nil && yellow(wb.marked[i]), yellow(wb.pieces[i]))
		}
	}
	for k, cells := range wb.markedEdges {
		if k[:min(len(k), 5)] == "view." {
			continue // the 3D view's own box: never a card's
		}
		for i, m := range cells {
			if m == nil || !yellow(m) {
				t.Errorf("edge %q cell %d not yellow", k, i)
			}
		}
	}
}
