package engine

import (
	"path/filepath"
	"slices"
	"testing"

	"gobox/adventure"
	"gobox/game"
	"gobox/rules"
	"gobox/rules/xxvc"
)

// br1ShopList1 is itemListEntries of list 1 (gdconsts.pas itemListStarts[1] = 1, itemListCounts[1] = 16), in the
// list's order; the shop shows the newest pushed first (eclvm.pas EclShop pushes each at the head).
var br1ShopList1 = []rules.ItemID{1, 3, 4, 6, 7, 9, 10, 12, 14, 15, 20, 21, 28, 29, 30, 31}

// br1Tables are the tables the test's import wrote.
func br1Tables(t *testing.T, e *Engine) (tab xxvc.Tables) {
	t.Helper()
	if ok, err := adventure.ReadJSON(filepath.Join(e.adv, "data"), adventure.RuleTablesFile, &tab); !ok || err != nil {
		t.Fatal("tables:", ok, err)
	}
	return tab
}

// br1Credits is what the member holds (the script's PartyCredits is the selected member's).
func br1Credits(e *Engine) rules.Copper {
	return rules.Copper(e.purse(false).Copper)
}

// TestBR1Shop: Shop(1) at PriceMultiplier 16 opens the shop with list 1's items at their values, newest first (shop.pas
// ItemPrice, eclvm.pas EclShop); a buy takes the price and gives the item; a sale asks first, then pays half the value
// (viewchar.pas SellItem); Leave goes on with the script after the shop.
func TestBR1Shop(t *testing.T) {
	e := br1Run(t, 201, "    PriceMultiplier = 16\n    Shop(1)\n    Word4C31 = 77")
	tab := br1Tables(t, e)
	tr, ok := e.build().(game.Trading)
	if !ok || tr.Kind != game.TradeShop || !tr.Buys {
		t.Fatalf("state %T %+v", e.build(), e.build())
	}
	var want []rules.ItemID
	for _, id := range slices.Backward(br1ShopList1) {
		if _, ok := tab.Items[id]; ok {
			want = append(want, id)
		}
	}
	if len(want) == 0 || len(tr.Goods) != len(want) {
		t.Fatalf("%d goods, want %d", len(tr.Goods), len(want))
	}
	for i, g := range tr.Goods {
		if kind := e.itemCard(rules.NoChar, g.Handle).Kind; kind != want[i] {
			t.Errorf("good %d is template %d, want %d", i, kind, want[i])
		}
		if v := max(tab.Items[want[i]].Value, 1); g.Price != int(v) {
			t.Errorf("good %d (%s) costs %d, want its value %d", i, g.Name, g.Price, v)
		}
	}
	if br1Cell(t, e, "Word4C31") == 77 {
		t.Fatal("the script ran on under the shop")
	}

	good := tr.Goods[0]
	before, n := br1Credits(e), len(e.listedItems(e.partyIDs()[0]))
	rules1 := rules.Copper(good.Price)
	if before < rules1 {
		if _, err := e.ruleset().Resolve(nil, rules.SetCredits{Who: e.selected(), Total: rules1 + 1000}, nil); err != nil {
			t.Fatal(err)
		}
		before = br1Credits(e)
	}
	if _, err := e.doNow(game.Buy{Member: 0, Item: good.Handle}); err != nil {
		t.Fatal("buy:", err)
	}
	if got := br1Credits(e); got != before-rules1 {
		t.Errorf("credits %d after buying for %d from %d", got, good.Price, before)
	}
	items := e.listedItems(e.partyIDs()[0])
	if len(items) != n+1 {
		t.Fatalf("carries %d items, was %d", len(items), n)
	}
	bought := items[len(items)-1]
	card := e.itemCard(e.partyIDs()[0], bought)
	if card.Kind != want[0] {
		t.Errorf("carries template %d, bought %d", card.Kind, want[0])
	}

	// a sale: the shop's offer is asked first, half the value (times the stack) is paid
	paid := br1Credits(e)
	sale := rules.Copper(max(tab.Items[want[0]].Value, 0)/2) * rules.Copper(max(card.Count, 1))
	if tab.Items[want[0]].NotJoinable {
		sale = rules.Copper(max(tab.Items[want[0]].Value, 0) / 2)
	}
	if _, err := e.doNow(game.Sell{Member: 0, Item: bought}); err != nil {
		t.Fatal("sell:", err)
	}
	if q := commonOf(e.build()).Ask; q != nil { // the front asks (its sell question with the published Sale); the engine does not ask again
		t.Fatalf("a second question after the front's sale: %+v", q)
	}
	if got := br1Credits(e); got != paid+sale || len(e.listedItems(e.partyIDs()[0])) != n {
		t.Errorf("after the sale: credits %d, want %d; %d items", got, paid+sale, len(e.listedItems(e.partyIDs()[0])))
	}

	if _, err := e.doNow(game.Leave{}); err != nil {
		t.Fatal("leave:", err)
	}
	if br1Cell(t, e, "Word4C31") != 77 {
		t.Error("the script did not go on after the shop")
	}
}

// TestBR1ShopRefusesAndReloads: short of credits a buy changes nothing and says why (shop.pas "Not enough Money."); a missile
// weapon bought is reloaded with a fill at ammoPrices a clip (shop.pas ShopAmmoMenu: 10 rounds a clip, 250 held), offered only
// while the shop is open and the weapon is not full.
func TestBR1ShopRefusesAndReloads(t *testing.T) {
	e := br1Run(t, 201, "    PriceMultiplier = 16\n    Shop(16)")
	tab := br1Tables(t, e)
	tr := e.build().(game.Trading)
	who := e.partyIDs()[0]
	var gun *game.Item
	for i, g := range tr.Goods {
		tm, ty := tab.Items[e.itemCard(rules.NoChar, g.Handle).Kind], xxvc.ItemType{}
		if int(tm.Type) < len(tab.ItemTypes) {
			ty = tab.ItemTypes[tm.Type]
		}
		if ty.Slot == xxvc.SlotWeapon && ty.WeaponClass != xxvc.Melee && ty.WeaponClass != xxvc.Grenade && ty.WeaponClass != xxvc.RocketLauncher &&
			ty.WeaponClass != xxvc.Plasma && tm.NotJoinable && tab.AmmoPrices[tm.Type] > 0 {
			gun = &tr.Goods[i]
			break
		}
	}
	if gun == nil {
		t.Fatal("list 16 has no missile weapon")
	}
	set := func(n rules.Copper) {
		t.Helper()
		if _, err := e.ruleset().Resolve(nil, rules.SetCredits{Who: who, Total: n}, nil); err != nil {
			t.Fatal(err)
		}
	}
	set(rules.Copper(gun.Price) - 1)
	n := len(e.listedItems(who))
	if _, err := e.doNow(game.Buy{Member: 0, Item: gun.Handle}); err == nil {
		t.Error("bought a weapon one credit short")
	}
	if len(e.listedItems(who)) != n || br1Credits(e) != rules.Copper(gun.Price)-1 || e.ext.note != "Not enough Money." {
		t.Errorf("a refused buy changed things: %d items, %d credits, note %q", len(e.listedItems(who)), br1Credits(e), e.ext.note)
	}
	set(rules.Copper(gun.Price) + 100000)
	if _, err := e.doNow(game.Buy{Member: 0, Item: gun.Handle}); err != nil {
		t.Fatal(err)
	}
	items := e.listedItems(who)
	h := items[len(items)-1]
	card := e.itemCard(who, h)
	ty := tab.Items[card.Kind].Type
	capacity := int32(250)
	clips := rules.Copper((capacity - card.Count + 9) / 10)
	if !slices.Contains(commonOf(e.build()).Allowed.Plain, game.Action(game.Reload{Member: 0, Item: h, Fill: true})) {
		t.Fatalf("a fill is not offered for a weapon with %d of %d rounds", card.Count, capacity)
	}
	paid := br1Credits(e)
	if _, err := e.doNow(game.Reload{Member: 0, Item: h, Fill: true}); err != nil {
		t.Fatal(err)
	}
	if got := e.itemCard(who, h).Count; got != capacity {
		t.Errorf("%d rounds after a fill, want %d", got, capacity)
	}
	if got, want := br1Credits(e), paid-clips*tab.AmmoPrices[ty]; got != want {
		t.Errorf("credits %d after the fill, want %d", got, want)
	}
	if _, err := e.doNow(game.Reload{Member: 0, Item: h, Fill: true}); err == nil {
		t.Error("a full weapon was reloaded")
	}
}
