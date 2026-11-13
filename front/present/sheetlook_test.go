package present

import (
	"image"
	"strings"
	"testing"

	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
)

// TestSheetDraftPage: the sheet drawn while a character is being made is the
// draft's page, not the selected member's.
func TestSheetDraftPage(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	var drawn strings.Builder
	g := func(c byte) [8]byte { drawn.WriteByte(c); return [8]byte{} }
	note := Note{Glyph: g, Font: func(c byte, _ bool) [8]byte { return g(c) }}
	st := play.Assembling{Common: play.Common{Party: []game.Member{{Name: "A", HP: 5, MaxHP: 5, Page: &game.SheetPage{Race: "DWARF"}}}},
		Draft: &play.Draft{Member: game.Member{Name: "Doc", Race: "ELF"}}}
	gm, _ := playGame(st)
	p := &Presenter{bars: gpu.NewImageFromImage(barArt())}
	p.SetEngine(gm)
	gm.r.sheet = 0
	p.drawSheet(dst, note)
	if s := drawn.String(); !strings.Contains(s, "ELF") || strings.Contains(s, "DWARF") {
		t.Fatalf("the sheet drew %q, want the draft's race", s)
	}
}

// TestLookMemberItems: in a fight the member whose look-only sheet is up is
// the one the items show, not the acting one.
func TestLookMemberItems(t *testing.T) {
	st := play.Fighting{Common: play.Common{Party: []game.Member{{Name: "A", Items: []game.Item{{Handle: 1, Name: "Mace"}}}, {Name: "B", Items: []game.Item{{Handle: 2, Name: "Sling"}}}}},
		Field: play.Battlefield{Size: image.Pt(6, 6), Turn: 1,
			Units: []game.Unit{{ID: 1, Member: 0, At: image.Pt(1, 1)}, {ID: 2, Member: -1, At: image.Pt(3, 1), Foe: true}}}}
	g, _ := playGame(st)
	g.r.items, g.r.lookSheet = game.ItemsOwn, 2
	in := g.r.inventoryOf(g.r.g.State())
	if in == nil || in.member != 1 || len(in.pool) != 1 || in.pool[0].Name != "Sling" {
		t.Fatalf("items of %+v, want B's Sling", in)
	}
}
