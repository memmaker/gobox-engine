package rulesxxvc

import (
	"fmt"
	"slices"
	"strings"

	"gobox/rules"
)

// takeItem is postcbt.pas LootTakeItemMenu's take: item i of the pile goes
// to who (Shop.GiveItemCopyToCreature) and, when it fits, leaves the pile
// (Took); over the item limit it stays (OverLoaded).
func (r *Rules) takeItem(s rules.Looting, a rules.TakeItem) (rules.Outcome, error) {
	c := r.chars[a.Who]
	if c == nil {
		return rules.Outcome{}, fmt.Errorf("TakeItem: no character %d", a.Who)
	}
	if a.Item < 0 || int(a.Item) >= len(s.Items) {
		return rules.Outcome{}, fmt.Errorf("TakeItem: no item %d in the pile", a.Item)
	}
	o, ok := r.giveCopy(a.Who, c, s.Items[a.Item])
	if ok {
		o.Effects = append(o.Effects, rules.Took{Who: a.Who, Item: a.Item})
	}
	return o, nil
}

// takeMoney is money.pas TakeMoneyMenu/TakeMoney: the credits asked for, up
// to the pile's (InputNumberMax caps the number), go to who.
func (r *Rules) takeMoney(s rules.Looting, a rules.TakeMoney) (rules.Outcome, error) {
	c := r.chars[a.Who]
	if c == nil {
		return rules.Outcome{}, fmt.Errorf("TakeMoney: no character %d", a.Who)
	}
	if a.Amount < 0 || a.Amount > s.Money {
		return rules.Outcome{}, fmt.Errorf("TakeMoney: %d of %d credits", a.Amount, s.Money)
	}
	if r.options.SharedPurse {
		r.purse += a.Amount
	} else {
		c.Coins += a.Amount
	}
	return rules.Outcome{}, nil
}

// lootGoods is the treasure's handles, in the order EndFight pushed them (handle order).
func (r *Rules) lootGoods() rules.Handles {
	hs := make(rules.Handles, 0, len(r.loot))
	for h := range r.loot {
		hs = append(hs, h)
	}
	slices.Sort(hs)
	return hs
}

// lootFact answers the treasure's Facts: the goods listed, and a treasure piece's card (named for no one).
func (r *Rules) lootFact(id rules.CharID, f rules.Fact) (rules.Value, bool, error) {
	switch f := f.(type) {
	case rules.ListedFact:
		if f.Goods {
			return r.lootGoods(), true, nil
		}
	case rules.ItemCardFact:
		if ref, ok := r.loot[f.Item]; ok && (r.chars[id] == nil || heldAt(r.chars[id], f.Item) < 0) {
			it, err := r.Item(ref)
			if err != nil {
				return nil, true, err
			}
			t := r.tables.Items[ref.ID]
			return rules.ItemCard{Item: f.Item, Name: strings.Join(it.NameWords, " "), Type: t.Type, Count: it.Count, Price: int32(r.shopPrice(ref)),
				Kind: ref.ID, Weight: t.Weight, Identified: it.Identified, Tab: string(it.Kind)}, true, nil
		}
	}
	return nil, false, nil
}

// bankLoot puts the pool's credits in the party's hands (money.pas: the pool is divided among the enabled
// members; under SharedPurse it is the one purse): the remainder goes to the first.
func (r *Rules) bankLoot() {
	pool := r.pool
	r.pool = 0
	if r.options.SharedPurse {
		r.purse += pool
		return
	}
	var heirs []*Character
	for _, id := range r.party {
		if c := r.chars[id]; c != nil && c.team() && c.Enabled {
			heirs = append(heirs, c)
		}
	}
	if len(heirs) == 0 {
		return
	}
	share := pool / rules.Copper(len(heirs))
	for _, c := range heirs {
		c.Coins += share
	}
	heirs[0].Coins += pool - share*rules.Copper(len(heirs))
}

// takeLoot is the treasure screen's Take: one piece (a.Item) or every piece in turn until one does not fit
// (a.All) goes to a.Who (Shop.GiveItemCopyToCreature); a taken piece leaves the pile.
func (r *Rules) takeLoot(a rules.TakeLoot) error {
	c := r.chars[a.Who]
	if c == nil {
		return fmt.Errorf("TakeLoot: no character %d", a.Who)
	}
	hs := rules.Handles{a.Item}
	if a.All {
		hs = r.lootGoods()
	}
	for _, h := range hs {
		ref, ok := r.loot[h]
		if !ok {
			return fmt.Errorf("TakeLoot: no treasure %d", h)
		}
		if _, ok := r.giveCopy(a.Who, c, ref); !ok {
			return nil // over the item limit: it stays
		}
		delete(r.loot, h)
	}
	return nil
}
