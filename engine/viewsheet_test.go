//go:build devviews

package engine

import (
	"testing"

	"gobox/game"
)

// TestViewSheet (BUG-CLASSES 36): mid-fight and in a shop each member's
// Sheet is the game's own page in full (Sex, Damage, saves ...), the same
// page the game's sheet shows; opening and closing it is gobox's.
func TestViewSheet(t *testing.T) { testViewSheet(t, testEngine) }

func testViewSheet(t *testing.T, mk func(*testing.T) *Engine) {
	full := func(what string, p *game.SheetPage) {
		t.Helper()
		if p == nil || p.Name == "" || p.Sex == "" || p.Race == "" || p.Class == "" || p.Damage == "" || p.Saves == [5]int{} || p.MaxHP == 0 {
			t.Fatalf("%s: sheet %+v", what, p)
		}
	}
	e := partyTurnSceneOn(t, mk(t))
	m := e.state().PartyIndex(e.member(e.cs.actor))
	fight := commonOf(e.build()).Party[m].Page
	full("fight", fight)
	if fight.Name != e.member(e.cs.actor).Name() {
		t.Errorf("fight sheet of %s, not the active %s", fight.Name, e.member(e.cs.actor).Name())
	}

	viewSheetShop(t, devPlayOn(t, mk(t), "shop"))
}

// viewSheetShop is TestViewSheet's shop half on e, in the shop dev scene.
func viewSheetShop(t *testing.T, e *Engine) {
	t.Helper()
	full := func(what string, p *game.SheetPage) {
		t.Helper()
		if p == nil || p.Name == "" || p.Sex == "" || p.Race == "" || p.Class == "" || p.Damage == "" || p.Saves == [5]int{} || p.MaxHP == 0 {
			t.Fatalf("%s: sheet %+v", what, p)
		}
	}
	if _, ok := e.build().(game.Trading); !ok {
		t.Fatalf("no shop: %T", e.build())
	}
	shop := commonOf(e.build()).Party[1].Page
	full("shop", shop)
	e.ext.creating = false
	e.selectMember(1)
	e.CharacterSheet() // the game's own sheet of the same member
	own := gameSheetPage(e.snap.sheet)
	if own == nil || own.Name != shop.Name || own.Sex != shop.Sex || own.Damage != shop.Damage || own.Saves != shop.Saves || own.THAC0 != shop.THAC0 || own.AC != shop.AC {
		t.Errorf("shop sheet %+v, the game's sheet %+v", shop, own)
	}
}
