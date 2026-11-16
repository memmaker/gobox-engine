package rulesxxvc

import (
	"fmt"
	"slices"
	"strings"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// maxCarriedItems is C_Item.MAX_CARRIED_ITEMS; maxStack is ViewChar.MAX_STACK.
const (
	maxCarriedItems = 16
	maxStack        = 100
)

// sameKind is JoinItem's test: two items whose templates agree on the name
// words, type, plus, save bonus, weight and specials.
func (r *Rules) sameKind(a, b Carried) bool {
	ta, tb := r.tables.Items[a.Ref.ID], r.tables.Items[b.Ref.ID]
	return ta.Words == tb.Words && ta.Type == tb.Type && plus(a, ta) == plus(b, tb) &&
		ta.SaveBonus == tb.SaveBonus && ta.Weight == tb.Weight && ta.Special == tb.Special
}

// joinItem is ViewChar.JoinItem for c.Items[i]: a joinable item takes in the
// others of its kind, in list order, up to maxStack a stack; a quantity of
// 0 counts as 1.
func (r *Rules) joinItem(c *Character, i int) {
	if r.tables.Items[c.Items[i].Ref.ID].NotJoinable {
		return
	}
	dest := c.Items[i].Slot
	at := func(slot rules.ItemSlot) *Carried {
		return &c.Items[slices.IndexFunc(c.Items, func(x Carried) bool { return x.Slot == slot })]
	}
	qty := max(at(dest).Ref.Count, 1)
	for _, slot := range slotsOf(c) { // the list as it was; counts read live, as the original walks it
		if slot == dest || !r.sameKind(*at(slot), *at(dest)) {
			continue
		}
		cur := *at(slot)
		n := max(cur.Ref.Count, 1)
		if qty+n <= maxStack {
			qty += n
			c.Items = slices.DeleteFunc(c.Items, func(x Carried) bool { return x.Slot == cur.Slot })
			continue
		}
		at(dest).Ref.Count = maxStack
		at(cur.Slot).Ref.Count = n - (maxStack - qty)
		qty, dest = at(cur.Slot).Ref.Count, cur.Slot
	}
	at(dest).Ref.Count = qty
}

func slotsOf(c *Character) []rules.ItemSlot {
	out := make([]rules.ItemSlot, len(c.Items))
	for i, x := range c.Items {
		out[i] = x.Slot
	}
	return out
}

// giveCopy is Shop.GiveItemCopyToCreature: a copy of ref goes to the end of
// c's list and joins its kind; over maxCarriedItems it is taken back
// ("OverLoaded") and ok is false; then RecalcStats.
func (r *Rules) giveCopy(who rules.CharID, c *Character, ref rules.ItemRef) (o rules.Outcome, ok bool) {
	slot := rules.ItemSlot(1)
	for _, x := range c.Items {
		slot = max(slot, x.Slot+1)
	}
	c.Items = append(c.Items, Carried{Slot: slot, Ref: ref})
	r.handOut(c)
	r.joinItem(c, len(c.Items)-1)
	ok = len(c.Items) <= maxCarriedItems
	if !ok {
		c.Items = slices.DeleteFunc(c.Items, func(x Carried) bool { return x.Slot == slot })
		o.Effects = append(o.Effects, rules.Narrated{Text: "OverLoaded"}, rules.Refused{Who: who, Why: rules.OverLoaded})
	}
	r.RecalcStats(c)
	return o, ok
}

// spend takes price from what id can spend: its own credits, or the party's
// one purse (the original's poolCredits covering what the credits lack).
func (r *Rules) spend(id rules.CharID, price rules.Copper) bool {
	if price > r.Purse(id) {
		return false
	}
	if r.options.SharedPurse {
		r.purse -= price
	} else {
		r.chars[id].Coins -= price
	}
	return true
}

// stockShop is EclShop: the pile (the loot's: availableItems) becomes list a.List's templates, a template that
// is no item (EMPTY_TEMPLATE, absent from the tables) left out, the newest pushed first (the lists' last entry
// heads the pile; the pile lists in handle order, so it is handed its handles last entry first). The terms are
// kept for the prices. Not a list of the tables: an error.
func (r *Rules) stockShop(a rules.StockShop) (rules.Outcome, error) {
	list, ok := r.tables.ShopLists[a.List]
	if !ok {
		return rules.Outcome{}, fmt.Errorf("StockShop: no list %d", a.List)
	}
	clear(r.loot)
	r.shopTerms = a.Terms
	for _, id := range slices.Backward(list) {
		if _, ok := r.tables.Items[id]; ok {
			r.lootHandle(rules.ItemRef{ID: id})
		}
	}
	return rules.Outcome{}, nil
}

// shopPrice is shop.pas ItemPrice after ShopBuyItemSelection raises a value
// of 0 to 1: the value times the zone's priceMultiplier sixteenths (0: the
// value itself).
func (r *Rules) shopPrice(ref rules.ItemRef) rules.Copper {
	v := rules.Copper(max(r.tables.Items[ref.ID].Value, 1))
	if r.shopTerms.PriceMultiplier == 0 {
		return v
	}
	return v * rules.Copper(r.shopTerms.PriceMultiplier) / 16
}

// buy is Shop.ShopBuyMenu's purchase of the pile's piece h at its shop price.
func (r *Rules) buy(who rules.CharID, h rules.ItemHandle) (rules.Outcome, error) {
	c := r.chars[who]
	if c == nil {
		return rules.Outcome{}, fmt.Errorf("Buy: no character %d", who)
	}
	ref, ok := r.loot[h]
	if !ok {
		return rules.Outcome{}, fmt.Errorf("Buy: no goods %d", h)
	}
	price := r.shopPrice(ref)
	if price > r.Purse(who) {
		return rules.Outcome{Effects: []rules.Effect{rules.Narrated{Text: "Not enough Money."}, rules.Refused{Who: who, Why: rules.NoMoney}}}, nil
	}
	if ref.Count == 0 {
		ref.Count = r.tables.Items[ref.ID].Quantity
	}
	o, ok := r.giveCopy(who, c, ref)
	if ok {
		r.spend(who, price)
		article := "a "
		if ref.Count > 0 {
			article = ""
		}
		o.Effects = append(o.Effects, rules.Traded{Who: who, Item: h}, rules.Narrated{Text: c.Name + " buys " + article + r.itemName(ref)})
	}
	return o, nil
}

// sellPrice is ViewChar.SellItem's offer: half the value, times the quantity
// of a joinable stack.
func (r *Rules) sellPrice(ci Carried) rules.Copper {
	t := r.tables.Items[ci.Ref.ID]
	price := rules.Copper(max(t.Value, 0) / 2)
	if ci.Ref.Count > 1 && !t.NotJoinable {
		price *= rules.Copper(ci.Ref.Count)
	}
	return price
}

// sell is ViewChar.SellItem: the shop's offer is asked first (Confirmed:
// sold, or a worthless item disposed of), the credits to the seller.
func (r *Rules) sell(who rules.CharID, h rules.ItemHandle, confirmed bool) (rules.Outcome, error) {
	c := r.chars[who]
	if c == nil {
		return rules.Outcome{}, fmt.Errorf("Sell: no character %d", who)
	}
	i := heldAt(c, h)
	if i < 0 {
		return rules.Outcome{}, fmt.Errorf("Sell: %d carries no item %d", who, h)
	}
	ci := c.Items[i]
	price, name := r.sellPrice(ci), r.itemName(ci.Ref)
	if !confirmed {
		q := fmt.Sprintf("I'll give you %d credits for your %s. Will you sell it?", price, name)
		switch {
		case price > 0 && ci.Ref.Count >= 2:
			q = fmt.Sprintf("I'll give you %d credits for your %s. Will you sell them?", price, name)
		case price == 0 && ci.Ref.Count < 2:
			q = "Your " + name + " is worthless to me. Shall I dispose of it for you?"
		case price == 0:
			q = "Your " + name + " are worthless. Would you like to discard them?"
		}
		return rules.Outcome{Confirm: q}, nil
	}
	text := "Sold!"
	switch {
	case price == 0 && ci.Ref.Count < 2:
		text = "I will give it to one of my sons."
	case price == 0:
		text = "Very well.  I'll recycle them."
	}
	c.Items = slices.Delete(c.Items, i, i+1) // RemoveItem
	if r.options.SharedPurse {
		r.purse += price
	} else {
		c.Coins += price // creditTotal, the seller's own
	}
	return rules.Outcome{Effects: []rules.Effect{rules.Traded{Who: who, Item: h, Sold: true}, rules.Narrated{Text: text}}}, nil
}

// itemName is the item's name words (PrintItemString).
// ponytail: no plus or quantity prefix; the engine names items on screen.
func (r *Rules) itemName(ref rules.ItemRef) string {
	it, _ := r.Item(ref)
	return strings.Join(it.NameWords, " ")
}

// Shop.ShopAmmoMenu's rounds (shop.pas): a rocket launcher or plasma weapon
// holds areaAmmoCapacity and is sold a round at a time; any other missile
// weapon ammoCapacity, sold in clips of clipRounds.
const (
	areaAmmoCapacity = 10
	ammoCapacity     = 250
	clipRounds       = 10
)

// reloadable is ShopAmmoMenu's list: a readied-slot missile weapon (not
// melee, not a grenade) that does not join (NOT_JOINABLE).
func (r *Rules) reloadable(ci Carried) (xxvc.ItemTemplate, bool) {
	t, ty, ok := r.itemType(ci)
	return t, ok && ty.Slot == xxvc.SlotWeapon && ty.WeaponClass != xxvc.Melee && ty.WeaponClass != xxvc.Grenade && t.NotJoinable
}

// ammoValues is ShopAmmoMenu's AmmoValues: the rounds a clip or a fill adds
// (never past the capacity), the weapon's capacity and the price, a clip's
// ammoPrices for each clip (a round each for the round-at-a-time classes).
func (r *Rules) ammoValues(ci Carried, t xxvc.ItemTemplate, kind rules.ReloadKind) (amount, capacity int32, cost rules.Copper) {
	ty := r.tables.ItemTypes[t.Type]
	single := ty.WeaponClass == xxvc.RocketLauncher || ty.WeaponClass == xxvc.Plasma
	amount, capacity = clipRounds, ammoCapacity
	if single {
		amount, capacity = 1, areaAmmoCapacity
	}
	if kind == rules.ReloadFill {
		amount = capacity
	}
	amount = min(amount, capacity-ci.Ref.Count)
	clips := amount
	if !single {
		clips = (amount + clipRounds - 1) / clipRounds
	}
	return amount, capacity, r.tables.AmmoPrices[t.Type] * rules.Copper(clips)
}

// reload is ShopAmmoMenu's 'C' (a clip) or 'F' (a fill): a full weapon is
// "Already full"; the price comes from who's credits, the pool covering what
// they lack, else "Not enough Money"; the rounds go in, up to the capacity.
func (r *Rules) reload(a rules.Reload) (rules.Outcome, error) {
	c := r.chars[a.Who]
	if c == nil {
		return rules.Outcome{}, fmt.Errorf("Reload: no character %d", a.Who)
	}
	if a.Kind != rules.ReloadClip && a.Kind != rules.ReloadFill {
		return rules.Outcome{}, fmt.Errorf("Reload: kind %d", a.Kind)
	}
	i := heldAt(c, a.Item)
	if i < 0 {
		return rules.Outcome{}, fmt.Errorf("Reload: %d carries no item %d", a.Who, a.Item)
	}
	t, ok := r.reloadable(c.Items[i])
	if !ok {
		return rules.Outcome{}, fmt.Errorf("Reload: item %d of %d is no missile weapon", a.Item, a.Who)
	}
	amount, capacity, cost := r.ammoValues(c.Items[i], t, a.Kind)
	say := func(s string) rules.Outcome { return rules.Outcome{Effects: []rules.Effect{rules.Narrated{Text: s}}} }
	if c.Items[i].Ref.Count >= capacity {
		return say("Already full"), nil
	}
	if !r.spend(a.Who, cost) {
		return say("Not enough Money"), nil
	}
	c.Items[i].Ref.Count = min(c.Items[i].Ref.Count+amount, capacity)
	if a.Kind == rules.ReloadClip {
		return say(c.Name + " purchases a clip"), nil
	}
	return say(c.Name + "'s weapon is reloaded"), nil
}
