package rulesxxvc

import (
	"errors"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// gearRules is two members and the items clip (1: joins), rock (2: NOT_JOINABLE) of the tables.
func gearRules(t *testing.T) (r *Rules, a, b rules.CharID) {
	t.Helper()
	r = newBR1()
	r.tables.Items = map[rules.ItemID]xxvc.ItemTemplate{1: {Type: 40, Weight: 1}, 2: {Type: 41, Weight: 1, NotJoinable: true}}
	a = r.Add(Character{Name: "A", Status: StOkay, Enabled: true, Items: []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 1, Count: 9}}, {Slot: 2, Ref: rules.ItemRef{ID: 2, Count: 1}}}})
	b = r.Add(Character{Name: "B", Status: StOkay, Enabled: true, Items: []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 1, Count: 5}}}})
	r.handOut(r.chars[a])
	r.handOut(r.chars[b])
	return r, a, b
}

func countOf(c *Character, id rules.ItemID) (n int32) {
	for _, x := range c.Items {
		if x.Ref.ID == id {
			n += x.Ref.Count
		}
	}
	return n
}

// TestGearOffers: the gear menu offers Trade to a player character (not in a fight), Halve with room, Join always.
func TestGearOffers(t *testing.T) {
	r, a, _ := gearRules(t)
	h := r.chars[a].Items[0].Handle
	got := map[string]bool{}
	for _, o := range r.Allowed(rules.Handling{Who: a, Item: h}).Offers {
		got[o.Name] = true
	}
	if !got["Trade"] || !got["Halve"] || !got["Join"] {
		t.Errorf("offers %v", got)
	}
	r.phase = rules.PhaseFight
	got = map[string]bool{}
	for _, o := range r.Allowed(rules.Handling{Who: a, Item: h}).Offers {
		got[o.Name] = true
	}
	if got["Trade"] || !got["Join"] {
		t.Errorf("in a fight: %v", got)
	}
	r.phase = rules.PhaseMenu
	r.chars[a].NonPlayer = true // a hireling in the party: not traded with
	if o := r.Allowed(rules.Handling{Who: a, Item: h}).Offers; len(o) == 0 || o[0].Name == "Trade" {
		t.Errorf("a hireling trades: %v", o)
	}
}

// TestGearHalveJoin: HalveItem splits off half, rounded down, unreadied; an item that does not join, or a single, is
// "Can't halve that"; JoinItem takes the kind back in, at most MAX_STACK a stack.
func TestGearHalveJoin(t *testing.T) {
	r, a, _ := gearRules(t)
	c := r.chars[a]
	if _, err := r.Resolve(nil, rules.Halve{Who: a, Item: c.Items[0].Handle}, nil); err != nil {
		t.Fatal(err)
	}
	if len(c.Items) != 3 || c.Items[0].Ref.Count != 5 || c.Items[1].Ref.Count != 4 || c.Items[1].Readied {
		t.Fatalf("halved 9: %+v", c.Items)
	}
	if _, err := r.Resolve(nil, rules.JoinStack{Who: a, Item: c.Items[0].Handle}, nil); err != nil {
		t.Fatal(err)
	}
	if len(c.Items) != 2 || countOf(c, 1) != 9 {
		t.Fatalf("joined: %+v", c.Items)
	}
	o, err := r.Resolve(nil, rules.Halve{Who: a, Item: c.Items[1].Handle}, nil) // the rock does not join
	if !errors.Is(err, rules.ErrNotNow) || len(o.Effects) != 1 || o.Effects[0] != rules.Effect(rules.Narrated{Text: gearCannotHalve}) {
		t.Errorf("halving a rock: %v %v", err, o.Effects)
	}
	c.Items = append(c.Items, Carried{Slot: 3, Ref: rules.ItemRef{ID: 1, Count: 95}})
	r.handOut(c)
	if _, err := r.Resolve(nil, rules.JoinStack{Who: a, Item: c.Items[0].Handle}, nil); err != nil {
		t.Fatal(err)
	}
	if c.Items[0].Ref.Count != maxStack || countOf(c, 1) != 104 {
		t.Errorf("a stack holds at most %d: %+v", maxStack, c.Items)
	}
}

// TestGearGive: ItemTradeMenu gives the item to the other, who joins it into his kind; a readied item is refused
// ("Must be unreadied"); a receiver with MAX_CARRIED_ITEMS items is "Overloaded" even for an item that would join.
func TestGearGive(t *testing.T) {
	r, a, b := gearRules(t)
	ca, cb := r.chars[a], r.chars[b]
	if _, err := r.Resolve(nil, rules.GiveItem{From: a, To: b, Item: ca.Items[0].Handle}, nil); err != nil {
		t.Fatal(err)
	}
	if len(ca.Items) != 1 || len(cb.Items) != 1 || cb.Items[0].Ref.Count != 14 {
		t.Fatalf("given: giver %+v receiver %+v", ca.Items, cb.Items)
	}
	ca.Items[0].Readied = true
	o, err := r.Resolve(nil, rules.GiveItem{From: a, To: b, Item: ca.Items[0].Handle}, nil)
	if !errors.Is(err, rules.ErrNotNow) || o.Effects[0] != rules.Effect(rules.Narrated{Text: gearUnready}) {
		t.Errorf("a readied item: %v %v", err, o.Effects)
	}
	ca.Items[0].Readied = false
	for i := len(cb.Items); i < maxCarriedItems; i++ {
		cb.Items = append(cb.Items, Carried{Slot: rules.ItemSlot(i + 1), Ref: rules.ItemRef{ID: 2, Count: 1}})
	}
	r.handOut(cb)
	o, err = r.Resolve(nil, rules.GiveItem{From: a, To: b, Item: ca.Items[0].Handle}, nil)
	if !errors.Is(err, rules.ErrNotNow) || o.Effects[0] != rules.Effect(rules.Narrated{Text: gearOverloaded}) || len(ca.Items) != 1 {
		t.Errorf("a full receiver: %v %v", err, o.Effects)
	}
}
