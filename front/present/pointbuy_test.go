package present

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gobox/dax"
	"gobox/front"
	"gobox/gpu"
	"gobox/input"
	"gobox/ui/fonts"
)

// pointsGame is the engine as the point buy asks it: the scores it was handed.
type pointsGame struct {
	front.Game
	scores  []int
	refused string
	cr      *front.Creation
}

func (g *pointsGame) CreateCharacter() error        { return nil }
func (g *pointsGame) PointBuy(s []int) error        { g.scores = s; return nil }
func (g *pointsGame) Refuse(s string)               { g.refused = s }
func (g *pointsGame) CreationView() *front.Creation { return g.cr }
func (g *pointsGame) Resets() int                   { return 0 }

func buyView(names []string) front.View {
	cr := &front.Creation{PointBuy: true, StatNames: names, Name: front.CellRect{X1: 1, Y1: 1, X2: 4, Y2: 1}}
	for i := range names {
		cr.Stats = append(cr.Stats, front.CellRect{X1: 1, Y1: 7 + i, X2: 8, Y2: 7 + i})
	}
	return front.View{Creation: cr}
}

// A new character is asked Roll or Point Buy; the point buy raises and
// lowers within the budget, and Done hands the scores to the game.
func TestPointBuy(t *testing.T) {
	for _, tc := range []struct {
		name  string
		stats []string
		font  func() []byte
	}{
		{"pool", []string{"STR", "INT", "WIS", "DEX", "CON", "CHA"}, func() []byte { return fonts.Get(theme.Font) }},
		{"br", []string{"STR", "DEX", "CON", "INT", "WIS", "CHA", "TECH"}, brBits},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := &pointsGame{}
			p := &Presenter{}
			p.SetEngine(g)
			dst := gpu.NewImage(640, 400)
			defer dst.Deallocate()
			bits := tc.font()
			gl := func(ch byte) [8]byte { r, _ := fonts.Glyph(bits, ch, true); return r }
			note := Note{Glyph: gl, Font: func(ch byte, text bool) [8]byte { r, _ := fonts.Glyph(bits, ch, text); return r }}
			v := buyView(tc.stats)
			g.cr = v.Creation
			p.send(front.ActCreate, 0)
			p.Draw(dst, note, v)
			if p.scr.sub != scrHow {
				t.Fatalf("sub %d, want the Roll or Point Buy question", p.scr.sub)
			}
			p.Key('P', v)
			if p.scr.sub != scrBuy || len(p.scr.buy) != len(tc.stats) {
				t.Fatalf("sub %d buy %v", p.scr.sub, p.scr.buy)
			}
			right, left, down := keyOf(input.Right), keyOf(input.Left), keyOf(input.Down)
			for range 12 { // STR to 18 (16 points), no further
				p.Key(right, v)
			}
			p.Key(down, v)
			for range 8 { // INT to 16 (10), back to 15 (8)
				p.Key(right, v)
			}
			p.Key(left, v)
			p.Key(down, v)
			for range 10 { // the next to 14: 6 left over were 8, 2 more stop it at 13 (5)... budget 32-24=8: 14 costs 6, 15 costs 8
				p.Key(right, v)
			}
			want := append([]int{18, 15, 15}, slices.Repeat([]int{8}, len(tc.stats)-3)...)
			if !slices.Equal(p.scr.buy, want) {
				t.Fatalf("bought %v, want %v", p.scr.buy, want)
			}
			p.Key('-', v) // Lower: the row last highlighted
			want[2] = 14
			if !slices.Equal(p.scr.buy, want) {
				t.Fatalf("after Lower %v, want %v", p.scr.buy, want)
			}
			p.Draw(dst, note, v)
			saveBuyPNG(t, dst, tc.name)
			p.Key('D', v)
			for _, f := range p.Calls() {
				f(g)
			}
			if !slices.Equal(g.scores, want) || g.refused != "" {
				t.Fatalf("game got %v (%q), want %v", g.scores, g.refused, want)
			}
		})
	}
}

// Without the game's PointBuy no question is asked.
func TestPointBuyNotOffered(t *testing.T) {
	p := &Presenter{}
	g := &pointsGame{}
	p.SetEngine(g)
	dst := gpu.NewImage(640, 400)
	defer dst.Deallocate()
	v := buyView([]string{"STR"})
	v.Creation.PointBuy = false
	g.cr = v.Creation
	p.send(front.ActCreate, 0)
	p.Draw(dst, Note{Glyph: func(byte) [8]byte { return [8]byte{} }, Font: func(byte, bool) [8]byte { return [8]byte{} }}, v)
	if p.scr.sub != scrMain {
		t.Fatalf("sub %d without the capability", p.scr.sub)
	}
}

// brBits is BR2's own font from its DOS files (shipsheet_test.go's brFont),
// gobox's copy where they are not installed.
func brBits() []byte {
	home, _ := os.UserHomeDir()
	if l, err := dax.NewLoader(filepath.Join(home, "Projects", "buckrogers2-go", "data", "dos")); err == nil {
		if b := brFont(l); b != nil {
			return b
		}
	}
	return fonts.Get("Buck_Rogers_8x8D1")
}

// saveBuyPNG writes dst to POINTBUY_PNG/name.png when set.
func saveBuyPNG(t *testing.T, dst *gpu.Image, name string) {
	out := os.Getenv("POINTBUY_PNG")
	if out == "" {
		return
	}
	m := image.NewRGBA(image.Rect(0, 0, 640, 400))
	dst.ReadPixels(m.Pix)
	fh, err := os.Create(filepath.Join(out, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	png.Encode(fh, m)
}
