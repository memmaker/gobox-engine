package present

import (
	"image/color"
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
	if headOf(game.Member{Name: "Buck", IconBodyID: "combat_icon_000"}) == nil {
		t.Fatal("no face for an icon-only member")
	}
}

// TestInventoryListsReadied: a readied item is on the doll and in the
// list, marked readied, so the keys reach it.
func TestInventoryListsReadied(t *testing.T) {
	in := upOf(play.Exploring{Common: play.Common{Party: []game.Member{{Items: []game.Item{{Name: "Gun", Kind: "weapon", Slot: "weapon", Readied: true}}}}}}, game.ItemsOwn)
	if in.doll[0] == nil || len(in.pool) != 1 || !in.pool[0].Readied {
		t.Fatalf("doll %v pool %+v", in.doll[0], in.pool)
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
	k := draftPage(s.Creation, s.SheetPage)
	if k == nil || k.Damage != "1d8" || k.THAC0 != 17 || k.AC != 4 || k.Name != "Doc" || k.Stats[0] != 12 {
		t.Fatalf("%+v", k)
	}
}

// The sheet's name colour is gobox's: a draft's name is never black.
func TestDraftNameColour(t *testing.T) {
	p, _ := draftOn(&play.Draft{Member: game.Member{Name: "Doc"}})
	ink := p.ink()
	c := p.screenPage()
	if runs := c.lines()[1]; len(runs) == 0 || runs[0].Fg == (color.RGBA{}) || runs[0].Fg == ink[0] {
		t.Fatalf("black name: %+v", runs)
	}
}

func TestMinimums(t *testing.T) {
	if s := minimums(map[string]int{"str": 9, "int": 12}); s != "9 12 - - - -" {
		t.Fatal(s)
	}
}
