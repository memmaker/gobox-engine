package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

func TestLoot(t *testing.T) {
	s := game.State{Mode: game.ModeLoot, Selected: 1, Loot: &game.Loot{Items: []game.Item{{Name: "Gauntlets"}}},
		Party: []game.Member{{}, {Spells: []game.SpellRef{{Name: detectMagic, Memorized: true}}}}}
	if l := Loot(s); l == nil || !l.Take || !l.Detect {
		t.Errorf("loot %+v", l)
	}
	s.Loot.Items = nil
	if l := Loot(s); l == nil || l.Take || l.Detect {
		t.Errorf("no treasure: %+v", l)
	}
	if l := Loot(s); l.Status {
		t.Error("Status offered unasked")
	}
	s.Loot.Status = true
	if l := Loot(s); !l.Status {
		t.Error("Status not offered")
	}
}

// TestLootScreen: the treasure up opens the one loot screen (the inventory
// with the treasure, its money and options) without the game's View, and
// its acts are TakeLoot.
func TestLootScreen(t *testing.T) {
	s := game.State{Mode: game.ModeLoot, Party: []game.Member{{}},
		Loot: &game.Loot{Items: []game.Item{{Name: "Gauntlets"}}, Money: map[string]int{game.Gold: 120, game.Gems: 2}}}
	g := &Game{R: &loreRules{s: s}}
	if v := g.ViewOf(s); v.Inventory != nil || v.Ask == nil || v.Ask.Menu.Prompt != "You found some loot!" || v.Loot != nil {
		t.Fatalf("no found message first: %+v", v.Ask)
	}
	g.Answer(front.Action{Kind: front.ActOK})
	v := g.ViewOf(s)
	in := v.Inventory
	if in == nil || !in.Loot || len(in.Wares) != 1 || in.Treasure == nil || v.Loot != nil {
		t.Fatalf("no loot screen: %+v loot %+v", in, v.Loot)
	}
	if m := in.Treasure.Money; len(m) != 2 || m[0] != "120 gold" || m[1] != "2 gems" {
		t.Errorf("money %q", m)
	}
	s.Sheet = &game.Sheet{}
	if g.ViewOf(s).Inventory != nil {
		t.Error("loot screen over the sheet")
	}
	s.Mode = game.ModeExplore
	g.ViewOf(s)
	s.Mode, s.Sheet = game.ModeLoot, nil
	if g.ViewOf(s).Inventory != nil {
		t.Error("the next treasure was not announced")
	}
}

func TestLootDecisions(t *testing.T) {
	g := &Game{R: &loreRules{s: game.State{Mode: game.ModeLoot, Loot: &game.Loot{}}}}
	for a, want := range map[front.Action]game.Decision{
		{Kind: front.ActBuy, N: 2}: game.TakeLoot{Item: 2},
		{Kind: front.ActTakeMoney}: game.TakeLoot{Item: -1},
		{Kind: front.ActTakeAll}:   game.TakeLoot{Item: -1, All: true},
	} {
		if d, ok := g.decision(a); !ok || d != want {
			t.Errorf("%v: %#v", a.Kind, d)
		}
	}
}
