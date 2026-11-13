package present

import (
	"testing"

	"gobox/core"
	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

func TestLoot(t *testing.T) {
	s := shown{Mode: game.ModeLoot, Selected: 1, Loot: &game.Loot{Items: []game.Item{{Name: "Gauntlets"}}},
		Party: []game.Member{{}, {Spells: []game.SpellRef{{Name: detectMagic, Memorized: true}}}}}
	if l := Loot(s.Mode, s.Loot, s.Allowed, s.Party, s.Selected); l == nil || !l.Take || !l.Detect {
		t.Errorf("loot %+v", l)
	}
	s.Allowed = map[string]bool{"LootDetect": false}
	if l := Loot(s.Mode, s.Loot, s.Allowed, s.Party, s.Selected); l.Detect {
		t.Error("Detect offered against Allowed")
	}
	s.Allowed["LootDetect"], s.Party[1].Spells = true, nil
	if l := Loot(s.Mode, s.Loot, s.Allowed, s.Party, s.Selected); !l.Detect {
		t.Error("Detect not offered by Allowed")
	}
	s.Allowed = nil
	s.Loot.Items = nil
	if l := Loot(s.Mode, s.Loot, s.Allowed, s.Party, s.Selected); l == nil || l.Take || l.Detect {
		t.Errorf("no treasure: %+v", l)
	}
	if l := Loot(s.Mode, s.Loot, s.Allowed, s.Party, s.Selected); l.Status {
		t.Error("Status offered unasked")
	}
	s.Loot.Status = true
	if l := Loot(s.Mode, s.Loot, s.Allowed, s.Party, s.Selected); !l.Status {
		t.Error("Status not offered")
	}
}

// TestLootScreen: the treasure up opens the one loot screen (the inventory
// with the treasure, its money and options) without the game's View, and
// its acts are TakeLoot.
func TestLootScreen(t *testing.T) {
	treasure := play.Looting{Common: play.Common{Party: []game.Member{{}}}, Items: []game.Item{{Name: "Gauntlets"}}, Money: map[string]int{game.Gold: 120, game.Gems: 2}}
	announced := treasure
	announced.Announce = true
	f := &fake{st: announced}
	g := NewGame(f, game.Data{})
	if v := g.View(); v.Inventory != nil || v.Ask == nil || v.Ask.Menu.Prompt != "You found some loot!" || v.Loot != nil {
		t.Fatalf("no found message first: %+v", v.Ask)
	}
	g.Answer(front.Action{Kind: front.ActOK})
	if len(f.sent) != 1 || f.sent[0] != (play.Continue{}) {
		t.Fatalf("OK on the found message sent %v", f.sent)
	}
	f.st = treasure
	v := g.View()
	in := v.Inventory
	if in == nil || !in.Loot || len(in.Wares) != 1 || in.Treasure == nil || v.Loot != nil {
		t.Fatalf("no loot screen: %+v loot %+v", in, v.Loot)
	}
	if m := in.Treasure.Money; len(m) != 2 || m[0] != "120 gold" || m[1] != "2 gems" {
		t.Errorf("money %q", m)
	}
	sheet := treasure
	sheet.Party = []game.Member{{SheetPowers: []game.Entry{{Name: "Heal"}}}}
	f.st = sheet
	if g.View().Inventory != nil {
		t.Error("loot screen over the sheet")
	}
}

func TestLootDecisions(t *testing.T) {
	for a, want := range map[front.Action]play.Action{
		{Kind: front.ActBuy, Item: 7}: play.Take{Item: 7},
		{Kind: front.ActTakeMoney}:    play.Take{Item: core.NoItem},
		{Kind: front.ActTakeAll}:      play.Take{Item: core.NoItem, All: true},
	} {
		f := &fake{st: play.Looting{Items: []game.Item{{}, {}, {Handle: 7}}}}
		g := NewGame(f, game.Data{})
		g.Answer(a)
		if len(f.sent) != 1 || f.sent[0] != want {
			t.Errorf("%v: %#v", a.Kind, f.sent)
		}
	}
}
