package parity

import (
	"reflect"
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"gobox/game"
	"por/rules"
)

// sameLoot says classic's treasure screen equals por's.
func sameLoot(t *testing.T, name string, e *rules.Engine) {
	t.Helper()
	v := e.View()
	if c := classic.View(v, e.Rules().State()); !reflect.DeepEqual(c.Loot, v.Loot) {
		t.Errorf("%s: loot %+v, por %+v", name, c.Loot, v.Loot)
	}
}

func TestLootParity(t *testing.T) {
	e := engine(t, true, "loot")
	settled(e, front.Action{})
	if e.View().Loot == nil {
		t.Fatal("no loot")
	}
	sameLoot(t, "loot", e)
	if s := e.Rules().State(); s.Mode != game.ModeLoot || len(s.Loot.Items) != 3 {
		t.Fatalf("loot state %v %+v", s.Mode, s.Loot)
	}
	e.LootView(-1) // the treasure's inventory over it
	settled(e, front.Action{})
	sameLoot(t, "loot view", e)
	settled(e, front.Action{Kind: front.ActClose})
	sameLoot(t, "loot back", e)
	e.LootTake(-1)
	settled(e, front.Action{})
	sameLoot(t, "loot taken", e)
	for i := range e.Rules().State().Party { // who has Detect Magic memorized
		e.Rules().Apply(game.Select{N: i})
		settled(e, front.Action{})
		sameLoot(t, "loot lead", e)
	}
}

// TestPlaceState: the shop, temple and hall in State as por's View shows them.
func TestPlaceState(t *testing.T) {
	sh := engine(t, true, "shop")
	settled(sh, front.Action{})
	s, v := sh.Rules().State(), sh.View()
	if v.Shop == nil || s.Mode != game.ModeShop || s.Shop == nil || len(s.Shop.Goods) != 6 {
		t.Fatalf("shop: %v %+v", s.Mode, s.Shop)
	}
	sh.Browse(-1)
	v = settled(sh, front.Action{})
	for _, w := range v.Inventory.Wares { // the wares sorted, the goods in the engine's order
		if g := sh.Rules().State().Shop.Goods[w.Index]; g.Name != w.Name || g.Price != w.Price.Gold {
			t.Errorf("ware %d: %+v, por %+v", w.Index, g, w)
		}
	}

	tp := engine(t, true, "temple")
	settled(tp, front.Action{})
	s, v = tp.Rules().State(), tp.View()
	if v.Temple == nil || len(s.Temple) != len(v.Temple.Cures) {
		t.Fatalf("temple: %+v, por %+v", s.Temple, v.Temple)
	}
	for i, c := range v.Temple.Cures {
		if s.Temple[i].Name != c.Name || s.Temple[i].Price != c.Price { // Needs is checked when the temple switches
			t.Errorf("cure %d: %+v, por %+v", i, s.Temple[i], c)
		}
	}

	for _, scene := range []string{"hall", "hallpoor"} {
		h := engine(t, true, scene)
		settled(h, front.Action{})
		if err := h.OpenTraining(0); scene == "hallpoor" { // no Train Character without the fee
			if err == nil || h.Rules().State().Training != nil {
				t.Errorf("hallpoor: %v, training %+v", err, h.Rules().State().Training)
			}
			continue
		} else if err != nil {
			t.Fatalf("%s: %v, menu %v", scene, err, h.View().MainMenu)
		}
		settled(h, front.Action{})
		tr, s := h.TrainingView(), h.Rules().State()
		if tr == nil || s.Training == nil || s.Training.Cost != tr.Price {
			t.Fatalf("%s: training %+v, por %+v", scene, s.Training, tr)
		}
		ready := len(s.Training.Ready) > 0 && s.Training.Ready[0] == 0
		if ready != (tr.Why == "") || ready && len(tr.Becomes) == 0 {
			t.Errorf("%s: ready %v trains %v, por why %q becomes %v", scene, s.Training.Ready, s.Training.Trains, tr.Why, tr.Becomes)
		}
	}
}
