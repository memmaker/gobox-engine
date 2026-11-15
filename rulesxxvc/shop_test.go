package rulesxxvc

import (
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// Shop by hand (shop.pas, viewchar.pas): a bought clip of 60 joins the 50
// carried as 100 and a stack of 10 (MAX_STACK); short of money nothing
// changes; a 17th item is taken back (OverLoaded); prices are ItemPrice; selling asks first, then pays half the
// value times the stack.
func TestBuySell(t *testing.T) {
	r := New()
	r.Begin(rules.Campaign{Tables: xxvc.Tables{ItemTypes: []xxvc.ItemType{{}},
		Items: map[rules.ItemID]xxvc.ItemTemplate{1: {Value: 7, Quantity: 60}, 2: {NotJoinable: true, Value: 3}, 3: {}}}}, rules.Options{})
	who := r.Add(Character{Name: "w", Enabled: true, Coins: 100, Items: []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 1, Count: 50}}}})
	shop := rules.Shopping{Goods: []rules.ShopItem{{Handle: 11, Item: rules.ItemRef{ID: 1}}, {Handle: 12, Item: rules.ItemRef{ID: 2}}, {Handle: 13, Item: rules.ItemRef{ID: 3}}},
		Terms: rules.ShopTerms{PriceMultiplier: 92}} // ItemPrice: 7*92/16 = 40, 3*92/16 = 17, a value 0 as 1: 92/16 = 5
	if p := r.shopPrice(shop.Goods[2].Item, shop.Terms); p != 5 {
		t.Errorf("price of a value 0: %d", p)
	}
	if p := r.shopPrice(shop.Goods[0].Item, rules.ShopTerms{}); p != 7 {
		t.Errorf("multiplier 0: %d", p)
	}
	if _, err := r.Resolve(shop, rules.Buy{Who: who, Item: 11}, nil); err != nil {
		t.Fatal(err)
	}
	c := r.chars[who]
	if len(c.Items) != 2 || c.Items[0].Ref.Count != 100 || c.Items[1].Ref.Count != 10 || c.Coins != 60 {
		t.Fatalf("%+v coins %d", c.Items, c.Coins)
	}
	c.Coins = 39
	if r.Resolve(shop, rules.Buy{Who: who, Item: 11}, nil); len(c.Items) != 2 || c.Coins != 39 {
		t.Fatal("bought without the money")
	}
	c.Coins = 14*17 + 25
	for len(c.Items) < maxCarriedItems {
		r.Resolve(shop, rules.Buy{Who: who, Item: 12}, nil)
	}
	if o, _ := r.Resolve(shop, rules.Buy{Who: who, Item: 12}, nil); len(c.Items) != maxCarriedItems || c.Coins != 25 || len(o.Effects) != 2 || o.Effects[1] != (rules.Refused{Who: who, Why: rules.OverLoaded}) {
		t.Fatalf("overloaded: %d items, coins %d, %v", len(c.Items), c.Coins, o.Effects)
	}
	sell := rules.Sell{Who: who, Item: c.Items[0].Handle}
	if o, _ := r.Resolve(shop, sell, nil); o.Confirm == "" || c.Coins != 25 {
		t.Fatal("sold without asking")
	}
	if r.Resolve(shop, rules.Confirmed{Action: sell}, nil); c.Coins != 25+3*100 || len(c.Items) != maxCarriedItems-1 {
		t.Fatalf("coins %d", c.Coins)
	}
}

// ShopAmmoMenu by hand: a clip is 10 rounds at the type's ammoPrices, a fill
// tops up to 250 (2 clips for 15 rounds); a rocket launcher a round at a time
// up to 10; a full weapon is refused; short of money nothing changes.
func TestReload(t *testing.T) {
	r := New()
	r.Begin(rules.Campaign{Tables: xxvc.Tables{
		ItemTypes:  []xxvc.ItemType{{}, {Slot: xxvc.SlotWeapon, WeaponClass: xxvc.Laser}, {Slot: xxvc.SlotWeapon, WeaponClass: xxvc.RocketLauncher}},
		Items:      map[rules.ItemID]xxvc.ItemTemplate{1: {Type: 1, NotJoinable: true}, 2: {Type: 2, NotJoinable: true}},
		AmmoPrices: map[int32]rules.Copper{1: 7, 2: 50}}}, rules.Options{})
	who := r.Add(Character{Name: "w", Enabled: true, Coins: 100, Items: []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 1, Count: 225}}, {Slot: 2, Ref: rules.ItemRef{ID: 2, Count: 9}}}})
	c := r.chars[who]
	shop := rules.Shopping{}
	steps := []struct {
		a     rules.Reload
		count int32
		coins rules.Copper
	}{
		{rules.Reload{Who: who, Item: 1, Kind: rules.ReloadClip}, 235, 93},
		{rules.Reload{Who: who, Item: 1, Kind: rules.ReloadFill}, 250, 79},
		{rules.Reload{Who: who, Item: 1, Kind: rules.ReloadClip}, 250, 79}, // Already full
		{rules.Reload{Who: who, Item: 2, Kind: rules.ReloadFill}, 10, 29},
	}
	for i, s := range steps {
		if _, err := r.Resolve(shop, s.a, nil); err != nil {
			t.Fatal(err)
		}
		if got := c.Items[s.a.Item-1].Ref.Count; got != s.count || c.Coins != s.coins {
			t.Fatalf("step %d: %d rounds, %d coins", i, got, c.Coins)
		}
	}
	c.Items[1].Ref.Count, c.Coins = 5, 49
	if r.Resolve(shop, rules.Reload{Who: who, Item: 2, Kind: rules.ReloadClip}, nil); c.Items[1].Ref.Count != 5 || c.Coins != 49 {
		t.Fatal("reloaded without the money")
	}
}
