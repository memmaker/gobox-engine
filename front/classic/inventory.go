package classic

import (
	"slices"
	"strings"

	"gobox/front"
	"gobox/game"
)

// DollSlots are the paper doll's slots by Inventory.Doll index, by
// game.Item.Slot name (arrows and bolts share the Missiles slot, the
// launcher takes the Launcher slot when the weapon slot is taken).
var DollSlots = []string{"weapon", "shield", "armor", "gauntlets", "helm", "belt", "robe", "cloak", "boots", "ring", "ring2", "arrows", "launcher"}

// dollSlot is the Doll index for it, -1 for none.
func dollSlot(it game.Item, doll []*front.InvItem) int {
	switch it.Slot {
	case "bolts":
		return slices.Index(DollSlots, "arrows")
	case "weapon":
		if doll[0] != nil {
			return slices.Index(DollSlots, "launcher")
		}
	}
	return slices.Index(DollSlots, it.Slot)
}

// powerKind is the inventory's member power for a Decision: Cast the
// spells (Spell -1 the spell effects), LayOnHands Heal, Cure Cure.
func powerKind(d game.Decision) (front.ActKind, bool) {
	switch d := d.(type) {
	case game.Cast:
		if d.Spell < 0 {
			return front.ActMagic, true
		}
		return front.ActSpells, true
	case game.LayOnHands:
		return front.ActHeal, true
	case game.Cure:
		return front.ActCure, true
	}
	return 0, false
}

// itemKinds are the party inventory's tabs by game.Item.Kind (the rest: misc).
var itemKinds = map[string]front.ItemKind{"weapon": front.KindWeapon, "armor": front.KindArmor}

// lines are an item's Look text as lines (nil: none).
func lines(t string) []string {
	if t == "" {
		return nil
	}
	return strings.Split(t, "\n")
}

// invItem is it as the inventory shows it at index i: its icon's type and
// words, its tab.
func invItem(i int, it game.Item) front.InvItem {
	return front.InvItem{Index: i, Name: it.Name, Type: byte(it.Type), Words: it.Words, Count: it.Count,
		Kind: itemKinds[it.Kind]}
}

// slotted: it has a doll slot (it can be readied).
func slotted(it game.Item) bool { return it.Kind != "" && it.Kind != "misc" }

// Inventory is v (the game's; nil: not up) built: the selected
// member's items, readied ones on the doll, what it may do, and with
// the game's wares up the shop's goods or the treasure beside them.
func Inventory(v *front.Inventory, mode game.Mode, party []game.Member, sel, slots, used int, purse map[string]int, allowed map[string]bool, shop *game.Shop, loot *game.Loot) *front.Inventory {
	if v == nil || sel < 0 || sel >= len(party) {
		return v
	}
	m := party[sel]
	n := &front.Inventory{Member: sel, Cap: slots, Used: used, Doll: make([]*front.InvItem, len(DollSlots)),
		Use: allowed["UseItem"], Sell: allowed["Sell"], Reload: shop != nil && shop.Reloads, Switch: allowed["Select"],
		Trade: allowed["Trade"], Halve: allowed["Halve"], Join: allowed["Join"], PoolCoins: shop != nil && allowed["PoolCoins"]}
	n.AC, n.THAC0, n.Stats = m.AC, m.THAC0, m.THAC0 != 0
	for _, p := range m.Powers {
		if k, ok := powerKind(p.Do); ok && (k != front.ActSpells || !n.Sell) { // no spells in a shop
			n.Powers = append(n.Powers, front.Power{Name: p.Label, Kind: k})
		}
	}
	for i, it := range m.Items {
		x := invItem(i, it)
		x.Slot, x.Wrong, x.Use, x.Look, x.Scribe = slotted(it), it.Wrong, it.Usable, lines(it.Text), slices.ContainsFunc(it.Scroll, func(s game.ScrollSpell) bool { return s.Scribing })
		x.Price = front.Money{Gold: it.Sale[Coin], Copper: it.Sale[game.Copper]}
		if d := dollSlot(it, n.Doll); it.Readied && d >= 0 && n.Doll[d] == nil {
			y := x
			n.Doll[d] = &y
		}
		x.Readied = it.Readied // on the doll too: the list keeps every item, marked readied, so the keys reach it (Ready toggles)
		n.Pool = append(n.Pool, x)
	}
	if v.Wares == nil {
		return n
	}
	n.Gold, n.Copper, n.Wares = purse[Coin], purse[game.Copper], []front.InvItem{}
	var goods []game.Item
	switch {
	case loot != nil:
		goods, n.Loot, n.Treasure = loot.Items, true, Loot(mode, loot, allowed, party, sel)
	case shop != nil:
		goods = shop.Goods
	}
	for i, it := range goods {
		x := invItem(i, it)
		x.Slot, x.Wrong, x.Look, x.Price = slotted(it), it.Wrong, lines(it.Text), front.Money{Gold: it.Price}
		n.Wares = append(n.Wares, x)
	}
	return n
}

// appraise is the gems and jewelry to value, with the offer made (o).
func appraise(purse map[string]int, o []game.Offer) front.Appraise {
	a := front.Appraise{Gems: purse[game.Gems], Jewelry: purse[game.Jewelry]}
	if len(o) > 0 {
		a.Valued, a.Jewel, a.CanKeep = o[0].Value, o[0].Kind == game.Jewelry, o[0].CanKeep
	}
	return a
}

// Temple is t (the game's; nil: not up) built: the services for
// the selected member, needed or why not.
func Temple(t *front.Temple, services []game.Service, offers []game.Offer, purse map[string]int, allowed map[string]bool, party []game.Member, sel int) *front.Temple {
	if t == nil || services == nil {
		return t
	}
	v := &front.Temple{Gold: purse[Coin], Appraise: appraise(purse, offers),
		Money: front.Money3{Take: allowed["Take"], Pool: allowed["PoolCoins"], Share: allowed["LootShare"]}}
	for _, sv := range services {
		v.Cures = append(v.Cures, front.Cure{Name: sv.Name, Price: sv.Price, Not: sv.Not[sel],
			Needed: sel < 0 || sel >= len(party) || slices.Contains(sv.Needs, sel)})
	}
	return v
}

// Shop is h (the game's; nil: not up) built: what Id would tell
// about, its price, the last reveal.
func Shop(h *front.Shop, sh *game.Shop, purse map[string]int, allowed map[string]bool, party []game.Member) *front.Shop {
	if h == nil || sh == nil {
		return h
	}
	v := &front.Shop{IDPrice: sh.IdentifyPrice, Gold: purse[Coin], Pooled: sh.Pooled[Coin], Appraise: appraise(purse, sh.Offers), IdentifiedN: sh.Identifies,
		Pool: allowed["PoolCoins"], Take: allowed["Take"], Share: allowed["LootShare"]}
	item := func(r game.ItemRef) *game.Item {
		if r.Member >= 0 && r.Member < len(party) && r.Item >= 0 && r.Item < len(party[r.Member].Items) {
			return &party[r.Member].Items[r.Item]
		}
		return nil
	}
	for _, r := range sh.Unknown {
		if it := item(r); it != nil {
			x := invItem(0, *it)
			v.Unknown = append(v.Unknown, front.Unknown{Member: r.Member, Item: r.Item, Owner: party[r.Member].Name, Name: it.Name, Icon: &x})
		}
	}
	if r := sh.Identified; r != nil {
		if it := item(*r); it != nil {
			x := invItem(0, *it)
			x.Look = lines(it.Text)
			v.Identified = &x
		}
	}
	return v
}
