package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// TestPartyIconHead: a member with no portrait (Buck Rogers) shows their
// combat icon on the card, never an empty frame.
func TestPartyIconHead(t *testing.T) {
	old := Series
	defer func() { Series = old }()
	Series = "buck-rogers"
	ms := Party([]game.Member{{Name: "Buck", IconBodyID: "combat_icon_000"}}, 0, nil)
	if len(ms) != 1 || ms[0].Head == nil {
		t.Fatal("no face for an icon-only member")
	}
}

// TestInventoryListsReadied: a readied item is on the doll and in the
// list, marked readied, so the keys reach it.
func TestInventoryListsReadied(t *testing.T) {
	s := shown{Party: []game.Member{{Items: []game.Item{{Name: "Gun", Kind: "weapon", Slot: "weapon", Readied: true}}}}}
	in := Inventory(&front.Inventory{}, s.Mode, s.Party, s.Selected, s.PartyCap, s.PoolItems, s.Purse, s.Allowed, s.Shop, s.Loot)
	if in.Doll[0] == nil || len(in.Pool) != 1 || !in.Pool[0].Readied {
		t.Fatalf("doll %v pool %+v", in.Doll[0], in.Pool)
	}
}

// A yes/no over the 3D view with its question in the message is still asked.
func TestYesNoOverExplore(t *testing.T) {
	q, ok := Ask(game.ModeExplore, &play.Ask{Kind: game.YesNo})
	if !ok || q == nil || q.Kind != front.AskConfirm || q.Prompt == "" {
		t.Fatalf("%+v", q)
	}
}

// The creation sheet keeps its draft layout and shows the rules' figures.
func TestDraftFigures(t *testing.T) {
	s := shown{Mode: game.ModeCreate, Creation: &game.Creation{Draft: game.Member{Name: "Doc", Stats: map[string]int{"str": 12}}},
		SheetPage: &game.SheetPage{AC: 4, THAC0: 17, Damage: "1d8"}}
	k := Sheet(front.View{}, s.Sheet, s.Portrait, s.IconBox, s.SheetPage, Drafting(s.Mode, s.Creation), s.Creation, OwnSubject(s.Mode, s.Creation, s.Party, s.Selected)).SheetPage
	if k == nil || k.Damage != "1d8" || k.THAC0 != 17 || k.AC != 4 || k.Name != "Doc" || k.Stats[0] != 12 {
		t.Fatalf("%+v", k)
	}
}

// The sheet's name colour is gobox's: a draft's name is never black.
func TestDraftNameColour(t *testing.T) {
	s := shown{Mode: game.ModeCreate, Creation: &game.Creation{Draft: game.Member{Name: "Doc"}}}
	if k := Sheet(front.View{}, s.Sheet, s.Portrait, s.IconBox, s.SheetPage, Drafting(s.Mode, s.Creation), s.Creation, OwnSubject(s.Mode, s.Creation, s.Party, s.Selected)).SheetPage; k.NameColour == 0 {
		t.Fatal("black name")
	}
}

func TestMinimums(t *testing.T) {
	if s := minimums(map[string]int{"str": 9, "int": 12}); s != "9 12 - - - -" {
		t.Fatal(s)
	}
}
