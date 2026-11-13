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
	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
	"gobox/input"
	"gobox/ui/fonts"
)

// buyDraft is a draft whose sheet has the abilities names (nil: STR .. CHA).
func buyDraft(names []string) *play.Draft {
	d := &play.Draft{}
	for _, n := range names {
		d.Stats = append(d.Stats, game.StatDef{Label: n})
	}
	return d
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
			p, f := draftOn(buyDraft(tc.stats))
			g := p.eng
			dst := gpu.NewImage(640, 400)
			defer dst.Deallocate()
			bits := tc.font()
			gl := func(ch byte) [8]byte { r, _ := fonts.Glyph(bits, ch, true); return r }
			note := Note{Glyph: gl, Font: func(ch byte, text bool) [8]byte { r, _ := fonts.Glyph(bits, ch, text); return r }}
			p.send(front.FrontActCreate, 0)
			p.Draw(dst, note)
			if p.scr.sub != scrHow {
				t.Fatalf("sub %d, want the Roll or Point Buy question", p.scr.sub)
			}
			p.Key('P')
			if p.scr.sub != scrBuy || len(p.scr.buy) != len(tc.stats) {
				t.Fatalf("sub %d buy %v", p.scr.sub, p.scr.buy)
			}
			right, left, down := keyOf(input.InputRight), keyOf(input.InputLeft), keyOf(input.Down)
			for range 12 { // STR to 18 (16 points), no further
				p.Key(right)
			}
			p.Key(down)
			for range 8 { // INT to 16 (10), back to 15 (8)
				p.Key(right)
			}
			p.Key(left)
			p.Key(down)
			for range 10 { // the next to 14: 6 left over were 8, 2 more stop it at 13 (5)... budget 32-24=8: 14 costs 6, 15 costs 8
				p.Key(right)
			}
			want := append([]int{18, 15, 15}, slices.Repeat([]int{8}, len(tc.stats)-3)...)
			if !slices.Equal(p.scr.buy, want) {
				t.Fatalf("bought %v, want %v", p.scr.buy, want)
			}
			p.Key('-') // Lower: the row last highlighted
			want[2] = 14
			if !slices.Equal(p.scr.buy, want) {
				t.Fatalf("after Lower %v, want %v", p.scr.buy, want)
			}
			p.Draw(dst, note)
			saveBuyPNG(t, dst, tc.name)
			p.Key('D')
			for _, c := range p.Calls() {
				c(g)
			}
			var got []int
			for _, a := range f.sent {
				if b, ok := a.(play.PointBuy); ok {
					got = b.Scores
				}
			}
			if !slices.Equal(got, want) || g.r.toast.Text != "" {
				t.Fatalf("game got %v (%q), want %v", got, g.r.toast.Text, want)
			}
		})
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
