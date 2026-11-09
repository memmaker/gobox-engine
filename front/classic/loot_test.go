package classic

import (
	"testing"

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
	treasure := play.Looting{Common: play.Common{Party: []play.Member{{}}}, Items: []game.Item{{Name: "Gauntlets"}}, Money: map[string]int{game.Gold: 120, game.Gems: 2}}
	f := &fake{st: treasure}
	g := NewGame(f, game.Data{})
	if v := g.View(); v.Inventory != nil || v.Ask == nil || v.Ask.Menu.Prompt != "You found some loot!" || v.Loot != nil {
		t.Fatalf("no found message first: %+v", v.Ask)
	}
	g.Answer(front.Action{Kind: front.ActOK})
	v := g.View()
	in := v.Inventory
	if in == nil || !in.Loot || len(in.Wares) != 1 || in.Treasure == nil || v.Loot != nil {
		t.Fatalf("no loot screen: %+v loot %+v", in, v.Loot)
	}
	if m := in.Treasure.Money; len(m) != 2 || m[0] != "120 gold" || m[1] != "2 gems" {
		t.Errorf("money %q", m)
	}
	asked := treasure // a question over the treasure (Buck 2's credits) does not announce it again
	asked.Ask = &play.Question{Text: "Sure?", Replies: []string{"Yes"}}
	f.st = asked
	g.View()
	f.st = treasure
	if g.View().Inventory == nil {
		t.Error("a question over the treasure announced it again")
	}
	sheet := treasure
	sheet.Party = []play.Member{{Powers: []game.Entry{{Name: "Heal"}}}}
	f.st = sheet
	if g.View().Inventory != nil {
		t.Error("loot screen over the sheet")
	}
	f.st = play.Exploring{}
	g.View()
	f.st = treasure
	if g.View().Inventory != nil {
		t.Error("the next treasure was not announced")
	}
}

func TestLootDecisions(t *testing.T) {
	for a, want := range map[front.Action]play.Action{
		{Kind: front.ActBuy, N: 2}: play.Take{Item: 2},
		{Kind: front.ActTakeMoney}: play.Take{Item: -1},
		{Kind: front.ActTakeAll}:   play.Take{Item: -1, All: true},
	} {
		f := &fake{st: play.Looting{}}
		g := NewGame(f, game.Data{})
		g.found = true
		g.Answer(a)
		if len(f.sent) != 1 || f.sent[0] != want {
			t.Errorf("%v: %#v", a.Kind, f.sent)
		}
	}
}
