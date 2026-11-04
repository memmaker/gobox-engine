package present

import (
	"fmt"
	"image"
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
	v := front.View{Party: party, Map: &front.MapSquares{}, Exploring: true, Adventure: true, Camp: &front.Camp{},
		Walls: &front.Walls{Key: "layout", Side: [][][4]int{{{1, 1, 1, 1}}}, Pool: map[int]string{1: "w"}}}
	for _, sz := range [][2]int{{1920, 1200}, {1388, 1190}, {1280, 800}} {
		dst := gpu.NewImage(sz[0], sz[1])
		for _, row := range []bool{false, true} {
			for _, off := range []bool{false, true} {
				Host.PartyRow, Host.MapOff = row, off
				p := &Presenter{}
				p.Draw(dst, note, v)
				var rs []image.Rectangle // the cards, then each menu's lines
				rs = append(rs, p.cards...)
				for _, ls := range p.campCtx {
					var u image.Rectangle
					for _, l := range ls {
						u = u.Union(l)
					}
					rs = append(rs, u)
				}
				if len(p.cards) != 8 || len(p.campCtx) != 8 {
					t.Fatalf("%v row %v: %d cards, %d menus", sz, row, len(p.cards), len(p.campCtx))
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
