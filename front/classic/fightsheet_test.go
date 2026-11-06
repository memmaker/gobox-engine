package classic

import (
	"testing"

	"gobox/game"
)

type fightRules struct {
	seqRules
	s game.State
}

func (r *fightRules) State() game.State { return r.s }

// TestFightSheetLookOnly: a fight's View is gobox's look-only sheet: the
// turn member's Member.Page (or one built from the Member), no game call,
// no sheet menu (Cast); Esc closes it without the game.
func TestFightSheetLookOnly(t *testing.T) {
	r := &fightRules{}
	r.s = game.State{Selected: 0, Party: []game.Member{{Name: "A"}, {Name: "B", Page: &game.SheetPage{Name: "B", Damage: "1d8", Sex: "Male"}}},
		Combat: &game.Combat{Turn: &game.Turn{Unit: 7}, Units: []game.Unit{{ID: 3, Member: 0}, {ID: 7, Member: 1}}}}
	g := NewGame(r, game.Data{})
	if g.OpenSheet(-1); len(r.got) != 0 {
		t.Fatalf("the game was called: %v", r.got)
	}
	v := g.View()
	if sp := v.SheetPage; sp == nil || sp.Damage != "1d8" || sp.Sex != "Male" || v.Sheet == nil || len(v.Sheet.Powers) != 0 {
		t.Fatalf("sheet %+v %+v", v.Sheet, v.SheetPage)
	}
	if g.ShowMember(0); g.View().SheetPage.Name != "A" {
		t.Fatal("no page built from the Member")
	}
	if g.CloseSheet(); len(r.got) != 0 || g.View().SheetPage != nil {
		t.Fatalf("close: %v", r.got)
	}
}

// TestFightItemsTabSheet: Tab on a fight's items (CloseItems false) closes
// them and opens the look-only sheet of the member they were, as Tab opens
// the sheet out of a fight.
func TestFightItemsTabSheet(t *testing.T) {
	r := &fightRules{}
	r.s = game.State{Selected: 1, Party: []game.Member{{Name: "A"}, {Name: "B"}},
		Combat: &game.Combat{Turn: &game.Turn{Unit: 3}, Units: []game.Unit{{ID: 3, Member: 0}}}}
	g := NewGame(r, game.Data{})
	g.CloseItems(false)
	if len(r.got) != 1 || r.got[0] != (game.CloseItems{All: true}) {
		t.Fatalf("sent %v", r.got)
	}
	if sp := g.View().SheetPage; sp == nil || sp.Name != "B" {
		t.Fatalf("sheet %+v", sp)
	}
}
