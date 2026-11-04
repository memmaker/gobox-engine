package parity

import (
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"gobox/front/present"
	"gobox/game"
)

// TestLootGate: the treasure screen is the presenter's menu over rule calls
// (moved from por's presentdata, now over classic.Game): Take opens the
// treasure beside the inventory, Exit there comes back to it, and leaving
// with treasure left asks first.
func TestLootGate(t *testing.T) {
	e := engine(t, true, "loot")
	g := gameOf(e)
	v := g.Settle()
	for i := 0; v.Loot == nil && i < 40; i++ {
		v = settled(e, front.Action{Kind: front.ActOK})
	}
	if v.Loot == nil || !v.Loot.Take {
		t.Fatalf("no treasure screen: pending %d ask %+v", v.Pending, v.Ask)
	}
	p := &present.Presenter{}
	p.SetEngine(g)
	if v = shownBy(g, p, front.Action{Kind: front.ActChoose, N: 1}); v.Inventory == nil {
		t.Fatal("Take opened no inventory")
	}
	if v = shownBy(g, p, front.Action{Kind: front.ActClose}); v.Inventory != nil || v.Loot == nil {
		t.Fatalf("Exit from the treasure: inventory %v loot %v", v.Inventory != nil, v.Loot)
	}
	if err := e.Rules().Apply(game.LootLeave{Sure: false}).Err; err == nil || err.Error() != "There is still treasure left. Leave it?" {
		t.Fatalf("leaving with treasure left: %v", err)
	}
	t.Skip("gap: the refusal carries no Ask (rules/gamecalls.go), so classic puts no Confirm to the presenter and Yes cannot be tried")
}

// reloaded is the game with n more resets.
type reloaded struct {
	*classic.Game
	n int
}

func (r *reloaded) Resets() int { return r.Game.Resets() + r.n }

// TestResetDropsOwn: a load or Back to Main Menu (the game's Resets
// changing) drops the presenter's own screens over the game, so none of
// them blanks the party menu or swallows its actions.
func TestResetDropsOwn(t *testing.T) {
	e := partyMenu(t)
	g := &reloaded{Game: gameOf(e)}
	p := &present.Presenter{}
	p.SetEngine(g)
	if v := shownBy(g.Game, p, front.Action{Kind: front.ActCreate}); v.Creation == nil {
		t.Fatal("no creation")
	}
	g.n++ // as if the game was loaded under it: the party menu is back
	e.EndCreation(false)
	settled(e, front.Action{})
	if v := p.Own(g.View()); v.MainMenu == nil || v.Creation != nil {
		t.Fatalf("own screens kept after a reset: menu %v creation %v", v.MainMenu != nil, v.Creation != nil)
	}
}

// TestSheetPNGAdapter: por's sheetpng golden (a PNG dropped on the sheet
// becomes the portrait) needs ActPortrait, which classic.Game does not answer.
func TestSheetPNGAdapter(t *testing.T) {
	t.Skip("gap: classic.Game answers no ActPortrait, so the sheetpng golden cannot be reached through it")
}
