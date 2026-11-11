package classic

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

type fightRules = fake

// TestFightSheetLookOnly: a fight's View is gobox's look-only sheet: the
// turn member's Member.Page (or one built from the Member), no game call,
// no sheet menu (Cast); Esc closes it without the game.
func TestFightSheetLookOnly(t *testing.T) {
	r := &fightRules{}
	r.st = play.Fighting{Common: play.Common{Party: []game.Member{{Name: "A"}, {Name: "B", Page: &game.SheetPage{Name: "B", Damage: "1d8", Sex: "Male"}}}},
		Field: play.Battlefield{Turn: 7, Units: []game.Unit{{ID: 3, Member: 0}, {ID: 7, Member: 1}}}}
	g := NewGame(r, game.Data{})
	if g.OpenSheet(-1); len(r.sent) != 0 {
		t.Fatalf("the game was called: %v", r.sent)
	}
	v := g.View()
	if sp := v.SheetPage; sp == nil || sp.Damage != "1d8" || sp.Sex != "Male" || v.Sheet == nil || len(v.Sheet.Powers) != 0 {
		t.Fatalf("sheet %+v %+v", v.Sheet, v.SheetPage)
	}
	if g.ShowMember(0); g.View().SheetPage.Name != "A" {
		t.Fatal("no page built from the Member")
	}
	if g.CloseSheet(); len(r.sent) != 0 || g.View().SheetPage != nil {
		t.Fatalf("close: %v", r.sent)
	}
}

// TestFightItemsTabSheet: Tab on a fight's items (CloseItems false) closes
// them and opens the look-only sheet of the member they were, as Tab opens
// the sheet out of a fight.
func TestFightItemsTabSheet(t *testing.T) {
	r := &fightRules{}
	r.st = play.Fighting{Common: play.Common{Selected: 1, Party: []game.Member{{Name: "A"}, {Name: "B"}}},
		Field: play.Battlefield{Turn: 3, Units: []game.Unit{{ID: 3, Member: 0}}}}
	g := NewGame(r, game.Data{})
	g.CloseItems(false)
	if sp := g.View().SheetPage; sp == nil || sp.Name != "B" {
		t.Fatalf("sheet %+v", sp)
	}
}
