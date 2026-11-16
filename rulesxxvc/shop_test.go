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
		Items:     map[rules.ItemID]xxvc.ItemTemplate{1: {Value: 7, Quantity: 60}, 2: {NotJoinable: true, Value: 3}, 3: {}},
		ShopLists: map[rules.ShopListID][]rules.ItemID{1: {1, 2, 3, 4}}}}, rules.Options{}) // 4 is no template (EMPTY_TEMPLATE): left out
	who := r.Add(Character{Name: "w", Enabled: true, Coins: 100, Items: []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 1, Count: 50}}}})
	if _, err := r.Resolve(nil, rules.StockShop{List: 2}, nil); err == nil {
		t.Error("a list the tables lack")
	}
	// ItemPrice: 7*92/16 = 40, 3*92/16 = 17, a value 0 as 1: 92/16 = 5
	if _, err := r.Resolve(nil, rules.StockShop{List: 1, Terms: rules.ShopTerms{PriceMultiplier: 92}}, nil); err != nil {
		t.Fatal(err)
	}
	goods := r.lootGoods() // the newest first: the list's last template heads the pile
	if len(goods) != 3 || r.loot[goods[0]].ID != 3 || r.loot[goods[2]].ID != 1 {
		t.Fatalf("pile %v", goods)
	}
	h1, h2, h3 := goods[2], goods[1], goods[0]
	for h, want := range map[rules.ItemHandle]rules.Copper{h1: 40, h2: 17, h3: 5} {
		if p := r.shopPrice(r.loot[h]); p != want {
			t.Errorf("price of %d: %d, want %d", h, p, want)
		}
	}
	if v, _, _ := r.lootFact(who, rules.ItemCardFact{Item: h1}); v.(rules.ItemCard).Price != 40 {
		t.Errorf("the card's price %v", v)
	}
	for mult, want := range map[int32]rules.Copper{0: 7, 16: 7, 8: 3, 32: 14, 1: 0} { // value 7: the value itself, sixteenths of it (1: 7/16 = 0)
		r.shopTerms.PriceMultiplier = mult
		if p := r.shopPrice(r.loot[h1]); p != want {
			t.Errorf("multiplier %d: %d, want %d", mult, p, want)
		}
	}
	r.shopTerms.PriceMultiplier = 92
	if _, err := r.Resolve(nil, rules.Buy{Who: who, Item: h1}, nil); err != nil {
		t.Fatal(err)
	}
	c := r.chars[who]
	if len(c.Items) != 2 || c.Items[0].Ref.Count != 100 || c.Items[1].Ref.Count != 10 || c.Coins != 60 {
		t.Fatalf("%+v coins %d", c.Items, c.Coins)
	}
	c.Coins = 39
	if o, _ := r.Resolve(nil, rules.Buy{Who: who, Item: h1}, nil); len(c.Items) != 2 || c.Coins != 39 || o.Effects[len(o.Effects)-1] != (rules.Refused{Who: who, Why: rules.NoMoney}) {
		t.Fatal("bought without the money")
	}
	c.Coins = 14*17 + 25
	for len(c.Items) < maxCarriedItems {
		r.Resolve(nil, rules.Buy{Who: who, Item: h2}, nil)
	}
	if o, _ := r.Resolve(nil, rules.Buy{Who: who, Item: h2}, nil); len(c.Items) != maxCarriedItems || c.Coins != 25 || len(o.Effects) != 2 || o.Effects[1] != (rules.Refused{Who: who, Why: rules.OverLoaded}) {
		t.Fatalf("overloaded: %d items, coins %d, %v", len(c.Items), c.Coins, o.Effects)
	}
	sell := rules.Sell{Who: who, Item: c.Items[0].Handle}
	if card, _ := r.itemCard(c, c.Items[0].Handle); card.Sale.Copper != 3*100 {
		t.Errorf("the card's sale %d", card.Sale)
	}
	if o, _ := r.Resolve(nil, sell, nil); o.Confirm == "" || c.Coins != 25 {
		t.Fatal("sold without asking")
	}
	if o, _ := r.Resolve(nil, rules.Confirmed{Action: sell}, nil); c.Coins != 25+3*100 || len(c.Items) != maxCarriedItems-1 || o.Effects[0] != (rules.Traded{Who: who, Item: sell.Item, Sold: true}) {
		t.Fatalf("coins %d", c.Coins)
	}
}

// TestStockShopHandles: a member's item never shares a handle with the pile's (one numbering).
func TestStockShopHandles(t *testing.T) {
	r := New()
	r.Begin(rules.Campaign{Tables: xxvc.Tables{Items: map[rules.ItemID]xxvc.ItemTemplate{1: {}, 2: {}}, ShopLists: map[rules.ShopListID][]rules.ItemID{3: {1, 2}}}}, rules.Options{})
	who := r.Add(Character{Name: "w", Enabled: true, Items: []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 1}}, {Slot: 2, Ref: rules.ItemRef{ID: 2}}}})
	r.Resolve(nil, rules.StockShop{List: 3}, nil)
	for _, h := range r.lootGoods() {
		if heldAt(r.chars[who], h) >= 0 {
			t.Errorf("pile piece %d is also carried", h)
		}
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
		if _, err := r.Resolve(nil, s.a, nil); err != nil {
			t.Fatal(err)
		}
		if got := c.Items[s.a.Item-1].Ref.Count; got != s.count || c.Coins != s.coins {
			t.Fatalf("step %d: %d rounds, %d coins", i, got, c.Coins)
		}
	}
	c.Items[1].Ref.Count, c.Coins = 5, 49
	if r.Resolve(nil, rules.Reload{Who: who, Item: 2, Kind: rules.ReloadClip}, nil); c.Items[1].Ref.Count != 5 || c.Coins != 49 {
		t.Fatal("reloaded without the money")
	}
}

// TestReloadOffers: the counter offers a clip and a fill of a missile weapon that is not full, at ammoPrices, in the shop's phase only.
func TestReloadOffers(t *testing.T) {
	r := New()
	r.Begin(rules.Campaign{Tables: xxvc.Tables{
		ItemTypes:  []xxvc.ItemType{{}, {Slot: xxvc.SlotWeapon, WeaponClass: xxvc.Laser}},
		Items:      map[rules.ItemID]xxvc.ItemTemplate{1: {Type: 1, NotJoinable: true}},
		AmmoPrices: map[int32]rules.Copper{1: 7}}}, rules.Options{})
	who := r.Add(Character{Name: "w", Enabled: true, Items: []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 1, Count: 225}}}})
	s := rules.Reloading{Who: who, Item: r.chars[who].Items[0].Handle}
	if len(r.Allowed(s).Offers) != 0 {
		t.Fatal("offered outside the shop")
	}
	r.Resolve(nil, rules.EnterPhase{Phase: rules.PhaseShop}, nil)
	got := r.Allowed(s).Offers
	if len(got) != 2 || got[0].Price != 7 || got[1].Price != 21 || got[0].Action != (rules.Reload{Who: who, Item: s.Item, Kind: rules.ReloadClip}) {
		t.Fatalf("offers %+v", got)
	}
	r.chars[who].Items[0].Ref.Count = 250
	if len(r.Allowed(s).Offers) != 0 {
		t.Fatal("offered for a full weapon")
	}
}
