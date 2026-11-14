package ui

import (
	"image"
	"image/color"
	"testing"

	"gobox/game"
	"gobox/ui/input"
)

// Every pool picker pages through all its names: each name is on exactly
// one page, every page's cells are framed, and the cell's picture is the
// name picked there.
func TestPoolPickerPagesAllNames(t *testing.T) {
	for kind, want := range map[int]string{PickHead: "portraithead", PickBody: "portraitbody", PickIconHead: "iconhead", PickWeapon: "iconbody"} {
		g, _ := playGame(game.Exploring{})
		if !g.OpenPicker(kind) || g.p.pool == nil {
			t.Fatalf("%s: no pool picker", want)
		}
		pp := g.p.pool
		v := g.PickerView(0)
		seen := 0
		for pg := 0; pg < v.Pages; pg++ {
			v := g.PickerView(pg)
			if v.Page != pg {
				t.Fatalf("%s: page %d laid out as %d", want, pg, v.Page)
			}
			for k, r := range v.Cells {
				i := pg*v.PerPage + k
				if r.X2 >= 39 || r.Y2 >= v.Bottom || (v.Pages > 1 && r.Y2 >= v.Prev.Y1) {
					t.Errorf("%s: cell %d %+v off the box", want, i, r)
				}
				if v.Portraits[i] == nil {
					t.Errorf("%s: %s has no picture", want, pp.names[i])
				}
				seen++
			}
		}
		t.Logf("%s: %d names, %d pages of %d", want, len(pp.names), v.Pages, v.PerPage)
		if seen != len(pp.names) || v.N != len(pp.names) {
			t.Errorf("%s: %d cells over the pages, N %d, %d names", want, seen, v.N, len(pp.names))
		}
	}
}

// The presenter's picker: PgDn from the first page reaches the last name,
// the cursor's cell is on the page shown, and Enter picks the cursor's name.
func TestPoolPickerPresenterPages(t *testing.T) {
	g, f := playGame(game.Exploring{})
	p := g.p
	p.SetEngine(g)
	run := func() {
		for _, c := range p.Calls() {
			c(g)
		}
	}
	p.send(ActPicker, PickHead)
	run()
	pk := p.picker(p.ink())
	if pk == nil || p.eng.p.pool == nil {
		t.Fatal("no pool picker")
	}
	names := p.eng.p.pool.names
	p.scr.cur = 0
	for range len(names) {
		pk = p.picker(p.ink())
		if i := p.scr.cur - pk.Page*pk.PerPage; i < 0 || i >= len(pk.Cells) {
			t.Fatalf("cursor %d not on page %d (%d cells)", p.scr.cur, pk.Page, len(pk.Cells))
		}
		p.pickerKey(keyOf(input.InputRight), pk)
		run()
	}
	pk = p.picker(p.ink())
	if p.scr.cur != len(names)-1 || pk.Page != pk.Pages-1 {
		t.Fatalf("cursor %d page %d/%d, %d names", p.scr.cur, pk.Page, pk.Pages, len(names))
	}
	p.pickerKey(keyOf(input.Enter), pk)
	run()
	if len(f.sent) == 0 {
		t.Fatal("nothing sent")
	}
	t.Logf("%T %+v", f.sent[len(f.sent)-1], f.sent[len(f.sent)-1])
}

// Every pool picture is fitted inside its own cell: a body's (11 by 6
// cells) is not squeezed into a head's 11 by 5 top.
func TestPoolPickerPicturesInTheirCells(t *testing.T) {
	for _, kind := range []int{PickHead, PickBody, PickIconHead, PickWeapon} {
		g, _ := playGame(game.Exploring{})
		g.OpenPicker(kind)
		v := g.PickerView(0)
		c := newCellPage(image.Point{}, [16]color.RGBA{}, nil)
		c.picker(v)
		if len(c.overs) != len(v.Cells) {
			t.Fatalf("kind %d: %d pictures for %d cells", kind, len(c.overs), len(v.Cells))
		}
		for k, r := range v.Cells {
			at, cell := c.overs[k].At, r.Pixels()
			if !at.In(cell) || at.Max.Y != cell.Max.Y || at.Min.X-cell.Min.X != cell.Max.X-at.Max.X {
				t.Errorf("kind %d cell %d: picture %v in cell %v, not bottom-centred inside it", kind, k, at, cell)
			}
		}
	}
}

// The pointer's cell is the drawn page's own (40x25), not the window grid
// the engine lays out on: the cell under it is the one framed and picked.
func TestPointerCellOnPageGrid(t *testing.T) {
	f := &Front{}
	f.Game = &classicGame{Game: NewGame(&fake{st: game.Exploring{}}, game.Data{}), p: &f.P}
	f.Game.r.grid = image.Pt(60, 33)
	f.pic = Rect{0, 0, 1280, 800}
	f.P.scr.at, f.P.scr.pw, f.P.scr.ph = f.pic, 320, 200
	if x, y, in := f.Cell(1279, 799); !in || x != 39 || y != 24 {
		t.Fatalf("bottom right pixel on cell %d, %d, want 39, 24", x, y)
	}
}
