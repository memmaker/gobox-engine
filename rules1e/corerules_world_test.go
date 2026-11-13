package rules1e

import (
	"gobox/rules"
	"testing"
)

func TestCheckWorld(t *testing.T) {
	r := &Rules{S: &RuleState{}}
	a := &Character{}
	a.HP, a.THAC0, a.Move, a.Class = 30, 0x29, 12, ClsClericRanger
	a.AC[0] = 0x3E
	a.Levels[ClsCleric] = 3
	b := &Character{}
	b.HP, b.Move, b.Class = 9, 6, ClsMage
	b.Levels[ClsMage] = 2
	p := rules.Party{r.ID(a), r.ID(b)}
	// a: (30+10+10+12)/10 = 6; b: (9+16)/10 = 2
	if v := r.Check(rules.PartyStrength{Party: p}, nil).Value; v != 8 {
		t.Errorf("PartyStrength %d, want 8", v)
	}
	if c := r.Check(rules.Census{Party: p, Fact: rules.FactMove}, nil); c.Low != 6 || c.High != 12 || c.Avg != 9 {
		t.Errorf("Census %+v", c)
	}
	if c := r.Check(rules.Census{Fact: rules.FactMove}, nil); c.Low != 0xFF || c.High != 0 || c.Avg != 0 {
		t.Errorf("empty Census %+v", c)
	}
	if !r.Check(rules.HasClass{Party: p, Class: CoreClassRanger}, nil).Yes || r.Check(rules.HasClass{Party: p[1:], Class: CoreClassRanger}, nil).Yes {
		t.Error("HasClass ranger")
	}
	b.Items = []*Item{{Type: 0x4B}}
	if w := r.Check(rules.HasItem{Whom: p, Item: ItemIDBrassKey}, nil); !w.Yes || w.Who != p[1] || r.Check(rules.HasItem{Whom: p[:1], Item: ItemIDBrassKey}, nil).Yes {
		t.Errorf("HasItem brass key %+v", w)
	}
}

func TestDuelClone(t *testing.T) {
	r := &Rules{S: &RuleState{}}
	a := &Character{}
	a.HP = 12
	a.Items = []*Item{{Type: 1}, {Type: 2}}
	r.S.Party = []*Character{a}
	if _, err := r.Resolve(nil, rules.DuelClone{Of: r.ID(a), Name: "ROLF", Icon: 7}, nil); err != nil {
		t.Fatal(err)
	}
	cl := r.S.Party[1]
	if cl.Name() != "ROLF" || cl.HP != 12 || cl.Side != SideMonsters || cl.Order != 7 || cl.Items[0].Type != 2 || cl.Items[0] == a.Items[1] {
		t.Errorf("clone %+v", cl)
	}
}

// TestMemoSlots: core.MemoSlots is the slots the rules use (PORMemory).
func TestMemoSlots(t *testing.T) {
	if MemorySlots != rules.MemoSlots {
		t.Error("core.MemoSlots")
	}
}

// TestLibraryLoot: TreasureOf's coins add to the loot; library items land at the head in turn, a random scroll holds mage spells only.
func TestLibraryLoot(t *testing.T) {
	tb := POR
	tb.Library = []*Item{{Type: 5, Hidden: 7}, nil}
	tb.Scroll = &RandomScroll{Type: 39, Last: 6, Mage: []SpellID{2, 5}}
	r := &Rules{S: &RuleState{Tables: &tb, Items: make([]ItemBase, 1)}}
	for range 2 {
		if _, err := r.Resolve(nil, rules.AddLoot{Loot: CoinsToPurse([7]uint16{4: 2, 5: 3})}, nil); err != nil {
			t.Fatal(err)
		}
	}
	if want := CoinsToPurse([7]uint16{4: 4, 5: 6}); r.S.Loot != want {
		t.Errorf("loot %+v, want %+v", r.S.Loot, want)
	}
	if _, err := r.Resolve(nil, rules.AddLibraryItems{IDs: []int32{1, 2, 9, rules.RandomScrollItem}, Known: true}, nil); err != nil {
		t.Fatal(err)
	}
	tr := r.S.Treasure
	if len(tr) != 2 || tr[0].Type != 39 || tr[1].Type != 5 || tr[1].Hidden != 0 || tb.Library[0].Hidden != 7 {
		t.Fatalf("treasure %+v", tr)
	}
	for _, s := range tr[0].Spells {
		if s != 2 && s != 5 {
			t.Errorf("scroll spells %v", tr[0].Spells)
		}
	}
	r.S.Tables = &POR
	if _, err := r.Resolve(nil, rules.AddLibraryItems{IDs: []int32{1}}, nil); err == nil {
		t.Error("POR has no library: want an error")
	}
}
