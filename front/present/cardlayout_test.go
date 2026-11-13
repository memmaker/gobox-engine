package present

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"gobox/front"
	"gobox/gpu"
)

// In camp every member's context menu is up at once: in the row and down
// the sides, with and without the minimap, no two cards or menus meet.
func TestCampCardsAndMenusApart(t *testing.T) {
	defer func(o Options, s, c int) { *Host, mapSide, mapCell = o, s, c }(*Host, mapSide, mapCell)
	mapSide, mapCell = 160, 10
	glyph := func(byte) [8]byte { return [8]byte{0x7e, 0x42, 0x42, 0x42, 0x42, 0x42, 0x7e, 0} }
	note := Note{Glyph: glyph, Font: func(byte, bool) [8]byte { return glyph(0) }}
	pw := []front.Power{{Name: "View"}, {Name: "Items"}, {Name: "Cast"}, {Name: "Memorize"}, {Name: "Drop"}}
	var party []front.Member
	for i := range 8 {
		party = append(party, front.Member{Name: fmt.Sprint("M", i), HP: 1, MaxHP: 1, Powers: pw})
	}
	v := front.View{Party: party, Map: &front.MapSquares{}, Exploring: true, Adventure: true, Camp: &front.Camp{}, Sweeping: true, Interact: true,
		Walls: &front.Walls{Key: "layout", Side: [][][4]int{{{1, 1, 1, 1}}}, Pool: map[int]string{1: "w"}}}
	for _, sz := range [][2]int{{1920, 1200}, {1388, 1190}, {1280, 800}} {
		dst := gpu.NewImage(sz[0], sz[1])
		for _, row := range []bool{false, true} {
			for _, off := range []bool{false, true} {
				Host.PartyRow, Host.MapOff = row, off
				p := campPresenter()
				p.Draw(dst, note, v)
				dump(t, dst, fmt.Sprintf("camp-%dx%d-row%v-off%v", sz[0], sz[1], row, off))
				var rs []image.Rectangle // the cards, then each menu's lines
				rs = append(rs, p.cards...)
				rs = append(rs, p.labelAt...)
				for _, ls := range p.campCtx {
					var u image.Rectangle
					for _, l := range ls {
						u = u.Union(l)
					}
					rs = append(rs, u)
				}
				if row { // a spell list from the first card's menu: above the cards, on screen
					w := v
					w.Camp, w.Ask = nil, &front.Ask{Kind: front.AskSpell, Entries: make([]front.Entry, 9)}
					for i := range w.Ask.Entries {
						w.Ask.Entries[i] = front.Entry{Text: fmt.Sprint("Spell ", i), Spell: 1 + i}
					}
					q := &Presenter{}
					q.Draw(dst, note, w)
					m := rs[len(p.cards)+2] // the first card's menu: the list takes its place (ctxBox)
					q.pickAt = image.Rect(m.Min.X, m.Min.Y, m.Max.X, m.Min.Y)
					q.Draw(dst, note, w)
					dump(t, dst, fmt.Sprintf("pick-%dx%d-off%v", sz[0], sz[1], off))
					var list image.Rectangle
					for _, l := range q.menuRects {
						list = list.Union(l)
					}
					if list.Empty() || !list.In(dst.Bounds()) {
						t.Errorf("%v mapOff %v: spell list %v off screen", sz, off, list)
					}
					for j, c := range q.cards {
						if list.Overlaps(c) {
							t.Errorf("%v mapOff %v: spell list %v meets card %d %v", sz, off, list, j, c)
						}
					}
				}
				if len(p.cards) != 8 || len(p.campCtx) != 8 {
					t.Fatalf("%v row %v: %d cards, %d menus", sz, row, len(p.cards), len(p.campCtx))
				}
				if len(p.labelAt) != 2 {
					t.Fatalf("%v: %d labels", sz, len(p.labelAt))
				}
				for i, a := range rs {
					if !a.In(dst.Bounds()) {
						t.Errorf("%v row %v mapOff %v: %d %v off screen", sz, row, off, i, a)
					}
					for j, b := range rs[:i] {
						if a.Overlaps(b) {
							t.Errorf("%v row %v mapOff %v: %d %v meets %d %v", sz, row, off, i, a, j, b)
						}
					}
				}
			}
		}
		dst.Deallocate()
	}
}

// dump writes dst as name.png into $LAYOUT_PNG, if set.
func dump(t *testing.T, dst *gpu.Image, name string) {
	d := os.Getenv("LAYOUT_PNG")
	if d == "" {
		return
	}
	b := dst.Bounds()
	m := image.NewRGBA(b)
	dst.ReadPixels(m.Pix)
	f, err := os.Create(filepath.Join(d, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	png.Encode(f, m)
}
