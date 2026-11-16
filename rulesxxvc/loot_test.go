package rulesxxvc

import (
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// LootTakeItemMenu: a taken item leaves the pile (Took); with 16 carried it
// stays (OverLoaded, no Took). TakeMoney pays the credits asked, never more
// than the pile holds.
func TestLoot(t *testing.T) {
	r := New()
	r.Begin(rules.Campaign{Tables: xxvc.Tables{ItemTypes: []xxvc.ItemType{{}},
		Items: map[rules.ItemID]xxvc.ItemTemplate{1: {NotJoinable: true}}}}, rules.Options{})
	who := r.Add(Character{Name: "w", Enabled: true})
	pile := rules.Looting{Items: []rules.ItemRef{{ID: 1, Count: 1}, {ID: 1, Count: 1}}, Money: 30}
	o, err := r.Resolve(pile, rules.TakeItem{Who: who, Item: 1}, nil)
	if err != nil || len(o.Effects) != 1 || o.Effects[0] != (rules.Took{Who: who, Item: 1}) {
		t.Fatalf("%v %v", o.Effects, err)
	}
	c := r.chars[who]
	for len(c.Items) < maxCarriedItems {
		r.Resolve(pile, rules.TakeItem{Who: who, Item: 0}, nil)
	}
	if o, _ := r.Resolve(pile, rules.TakeItem{Who: who, Item: 0}, nil); len(c.Items) != maxCarriedItems || o.Effects[len(o.Effects)-1] != (rules.Refused{Who: who, Why: rules.OverLoaded}) {
		t.Fatalf("overloaded: %v", o.Effects)
	}
	if _, err := r.Resolve(pile, rules.TakeMoney{Who: who, Amount: 20}, nil); err != nil || c.Coins != 20 {
		t.Fatalf("coins %d %v", c.Coins, err)
	}
	if _, err := r.Resolve(pile, rules.TakeMoney{Who: who, Amount: 31}, nil); err == nil {
		t.Fatal("took more than the pile")
	}
}

// TestTreasureOf (eclvm.pas EclTreasure): the credits are the party's at once, the templates join the pile
// (an empty template left out, the call's last heading it), a coin that is not credits is an error.
func TestTreasureOf(t *testing.T) {
	r := New()
	r.Begin(rules.Campaign{Tables: xxvc.Tables{Items: map[rules.ItemID]xxvc.ItemTemplate{1: {}, 2: {}}}}, rules.Options{SharedPurse: true})
	r.Add(Character{Name: "w", Enabled: true})
	if _, err := r.Resolve(nil, rules.AddLoot{Loot: rules.Purse{Copper: 250}}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Resolve(nil, rules.AddLibraryItems{IDs: []int32{1, 9, 2}}, nil); err != nil {
		t.Fatal(err)
	}
	if r.purse != 250 || r.pool != 0 {
		t.Errorf("purse %d pool %d, want 250 and 0", r.purse, r.pool)
	}
	goods := r.lootGoods()
	if len(goods) != 2 || r.loot[goods[0]].ID != 2 || r.loot[goods[1]].ID != 1 {
		t.Errorf("pile %v", goods)
	}
	if _, err := r.Resolve(nil, rules.AddLoot{Loot: rules.Purse{Gems: 1}}, nil); err == nil {
		t.Error("gems are no credits")
	}
}

// chargen.pas AddCharacterMenu: a character is refused by name only (a party member of that name), and the
// player characters ahead of the match are counted against six (ControlledBefore); there is no rule of its own.
func TestJoinChecks(t *testing.T) {
	r := New()
	r.Begin(rules.Campaign{Tables: xxvc.Tables{}}, rules.Options{})
	a, b, twin, hired := r.Add(Character{Name: "a"}), r.Add(Character{Name: "b"}), r.Add(Character{Name: "a"}), r.Add(Character{Name: "h", NonPlayer: true})
	for _, id := range []rules.CharID{a, hired, b} {
		if _, err := r.Resolve(nil, rules.JoinParty{Who: id}, nil); err != nil {
			t.Fatal(err)
		}
	}
	if !r.Check(rules.CanJoin{Who: b}, nil).Yes || r.Check(rules.CanJoin{Who: 99}, nil).Yes {
		t.Error("CanJoin: a known character may")
	}
	if !r.Check(rules.InParty{Who: twin}, nil).Yes || r.Check(rules.InParty{Who: rules.CharID(99)}, nil).Yes {
		t.Error("InParty: by name")
	}
	if got := r.Check(rules.ControlledBefore{Who: b}, nil).Value; got != 1 { // a; the hired one is no player character
		t.Errorf("ControlledBefore b: %d, want 1", got)
	}
	if got := r.Check(rules.ControlledBefore{Who: 99}, nil).Value; got != 2 {
		t.Errorf("ControlledBefore a stranger: %d, want 2", got)
	}
}
