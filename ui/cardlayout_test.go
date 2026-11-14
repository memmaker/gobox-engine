package ui

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"gobox/game"
	"gobox/gpu"
	"gobox/rules"
)

// In camp every member's context menu is up at once: in the row and down
// the sides, with and without the minimap, no two cards or menus meet.
func TestCampCardsAndMenusApart(t *testing.T) {
	defer func(o Options) { *Host = o }(*Host)
	glyph := func(byte) [8]byte { return [8]byte{0x7e, 0x42, 0x42, 0x42, 0x42, 0x42, 0x7e, 0} }
	note := Note{Glyph: glyph, Font: func(byte, bool) [8]byte { return glyph(0) }}
	var members []game.Member
	var allowed game.Allowed // each card's menu: View, Items, Cast, Memorize, Drop
	for i := range 8 {
		members = append(members, game.Member{Name: fmt.Sprint("M", i), HP: 1, MaxHP: 1})
		allowed.Plain = append(allowed.Plain, game.Ready{Member: i}, game.Cast{Member: i}, game.Memorize{Member: i}, game.DropMember{Member: i})
	}
	presenter := func() *Presenter { // camp over the 3D view and its minimap the state gave
		common := game.Common{Party: members, Allowed: allowed, Adventure: true, Where: game.Where{Searching: true}} // the Sweeping label
		g, f := playGame(game.Exploring{Common: common, Map: game.ExploreMap{Automap: game.Automap{Size: rules.Size{W: 16, H: 16}},
			Walls: game.MapWalls{Side: [][][4]int{{{1, 1, 1, 1}}}, Pool: map[int]string{1: "w"}}}})
		p := &Presenter{}
		p.SetEngine(g)
		p.walls()
		f.st = game.Camping{Common: common}
		return p
	}
	for _, sz := range [][2]int{{1920, 1200}, {1388, 1190}, {1280, 800}} {
		dst := gpu.NewImage(sz[0], sz[1])
		for _, row := range []bool{false, true} {
			for _, off := range []bool{false, true} {
				Host.PartyRow, Host.MapOff = row, off
				p := presenter()
				p.Draw(dst, note)
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
				if row { // a spell list: in the middle, on screen, clear of the cards
					spells := FrontAsk{Kind: AskSpell, Entries: make([]FrontEntry, 9)}
					for i := range spells.Entries {
						spells.Entries[i] = FrontEntry{Text: fmt.Sprint("Spell ", i), Spell: 1 + i}
					}
					q := presenter()
					q.q = &Question{Ask: spells}
					q.Draw(dst, note)
					dump(t, dst, fmt.Sprintf("pick-%dx%d-off%v", sz[0], sz[1], off))
					var list image.Rectangle
					for _, l := range q.menuRects {
						list = list.Union(l)
					}
					if list.Empty() || !list.In(dst.Bounds()) {
						t.Errorf("%v mapOff %v: spell list %v off screen", sz, off, list)
					}
					if c, mid := list.Min.Add(list.Max).Div(2), dst.Bounds().Max.Div(2); max(c.X-mid.X, mid.X-c.X) > list.Dx()/2 || max(c.Y-mid.Y, mid.Y-c.Y) > list.Dy() { // out of a fight: in the middle, as the other menus
						t.Errorf("%v mapOff %v: spell list %v not in the middle %v", sz, off, list, mid)
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
				if len(p.labelAt) != 1 {
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
