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
