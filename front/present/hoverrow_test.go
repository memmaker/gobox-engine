package present

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/gpu"
)

// TestHoverRowNotSelected: the row under the pointer is not filled as the
// selected one, so only one row looks selected.
func TestHoverRowNotSelected(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	note := Note{Glyph: func(byte) [8]byte { return [8]byte{} }, Font: func(byte, bool) [8]byte { return [8]byte{} }}
	v := front.View{Party: []front.Member{{Name: "Kade", Selected: true}}, Inventory: &front.Inventory{Member: 0, Pool: []front.InvItem{{Index: 0, Name: "Mace"}, {Index: 1, Name: "Sling"}}}}
	p := &Presenter{}
	p.Inventory(dst, note, v)
	var rows []image.Rectangle
	for _, h := range p.inv.lines {
		if !h.doll {
			rows = append(rows, h.r)
		}
	}
	if len(rows) < 2 || p.inv.sel != 0 {
		t.Fatalf("rows %v sel %d", rows, p.inv.sel)
	}
	mid := func(r image.Rectangle) image.Point { return image.Pt(r.Max.X-3, (r.Min.Y+r.Max.Y)/2) }
	p.inv.hover = mid(rows[1])
	p.Inventory(dst, note, v)
	blue := col(inkBlue)
	if c := p.inv.px.RGBAAt(mid(rows[0]).X, mid(rows[0]).Y); c != blue {
		t.Fatalf("selected row %v, want %v", c, blue)
	}
	if c := p.inv.px.RGBAAt(mid(rows[1]).X, mid(rows[1]).Y); c == blue {
		t.Fatal("the hovered row is filled as the selected one")
	}
}
