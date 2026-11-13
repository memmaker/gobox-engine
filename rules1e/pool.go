package rules1e

import (
	"gobox/rules"
	"slices"
	"strings"
)

// The party inventory (extra): the party shares one list of every item no
// member has readied, limited by slot count (a stack is one slot); item
// weight and encumbrance are gone. A member's own list (Character.Items)
// holds only their readied items. Saved games keep the format: the pool is
// written with the first member's items, unreadied, and ApplyItems moves
// unreadied items of party members into it again on loading.

// PoolSlots is the party inventory's size, PoolBag what a readied bag of
// holding adds (balanced against the original's 16 items a member could
// carry with only a minor encumbrance penalty).
const (
	PoolSlots = 60
	PoolBag   = 10
)

// partyMember: c is one of the player's party (not a monster or the clone
// the scripts add).
func (e *RuleState) partyMember(c *Character) bool {
	return c.Side == SideParty && slices.Contains(e.Party, c)
}

// PoolCap is the number of slots of the party inventory.
func (e *RuleState) PoolCap() int {
	for _, c := range e.Party {
		for _, it := range c.Items {
			if it.Readied != 0 && it.Words[0] == 0xBA { // bag of holding
				return PoolSlots + PoolBag
			}
		}
	}
	return PoolSlots
}

// Carried is everything c can reach: their readied items and, for a party
// member, the party inventory.
func (e *RuleState) Carried(c *Character) []*Item {
	if !e.partyMember(c) {
		return c.Items
	}
	return append(c.Items[:len(c.Items):len(c.Items)], e.Pool...)
}

// poolFull: the party inventory has no room for it (no stack of it to join).
func (e *RuleState) poolFull(it *Item) bool {
	if len(e.Pool) < e.PoolCap() {
		return false
	}
	if it.Count > 0 && it.Spells[0] < 2 {
		for _, p := range e.Pool {
			if sameStack(p, it) && int(p.Count)+int(it.Count) <= 0xFF {
				return false
			}
		}
	}
	return true
}

// stow moves a party member's unreadied items into the party inventory.
// Items that need no slot (potions, wands, scrolls, keys ...) are only
// readied when they have a slot on the paper doll (readyItem).
func (e *RuleState) stow(c *Character) {
	keep := c.Items[:0]
	var moved []*Item
	for _, it := range c.Items {
		if it.Readied != 0 && e.DollSlot(c, it) < 0 {
			it.Readied = 0
			if it.Spells[2] > 0x7F {
				e.itemAffect(it, c, 1)
			}
		}
		if it.Readied != 0 {
			keep = append(keep, it)
		} else {
			moved = append(moved, it)
		}
	}
	clear(c.Items[len(keep):])
	c.Items = keep
	for _, it := range moved {
		e.addToPool(it)
	}
}

// addToPool puts the item in the party inventory, joining a like stack.
// It goes in even when the pool is full: callers test poolFull first where
// the player may refuse (the overflow only stops further pick-ups).
func (e *RuleState) addToPool(it *Item) *Item {
	it.Readied = 0
	if it.Count > 0 && it.Spells[0] < 2 {
		for _, p := range e.Pool {
			if sameStack(p, it) && int(p.Count)+int(it.Count) <= 0xFF {
				p.Count += it.Count
				return p
			}
		}
	}
	e.Pool = append(e.Pool, it)
	return it
}

// takeFromPool moves an item of the party inventory into c's own list.
func (e *RuleState) takeFromPool(c *Character, it *Item) {
	if i := slices.Index(e.Pool, it); i >= 0 {
		e.Pool = slices.Delete(e.Pool, i, i+1)
		c.Items = append(c.Items, it)
	}
}

// DollSlot is the doll slot a readied item of c fills, -1 for an item
// that cannot be readied (no slot: it is used straight from the pool).
func (e *RuleState) DollSlot(c *Character, it *Item) int {
	switch s := e.readySlot(it.Type); {
	case it.Readied == ReadiedOffHand || e.offHand(c, it): // the off hand (extra)
		return SlotShield
	case e.isKind(it.Type, ItemIDArrow, ItemIDQuarrel): // the arrows' and quarrels' slots (11, 12)
		return rules.DollArrows
	case s <= 8:
		return s
	case s == 9:
		if c.Slots[slotRing1] == it || c.Slots[slotRing1] == nil {
			return rules.DollRing1
		}
		return rules.DollRing2
	}
	return -1
}

// InvFlags say what c's inventory offers: using items, selling them to a
// shop, showing another member.
func (e *RuleState) InvFlags(c *Character) (use, sell, switches bool) {
	hired := c.Control >= 0x80 && c.Okay != 0 && c.Status != StatusAnimated
	use = c.Okay != 0 && e.Area.NoMagic == 0 && (e.BeginFlag == 2 || e.BeginFlag == 3 || e.BeginFlag == 4 ||
		e.BeginFlag == 5 && c.Combat != nil && c.Combat.CanUse != 0)
	return use, !hired && e.BeginFlag == 1, e.BeginFlag != 5
}

// Listed is items in the order the inventory lists them (by kind, the
// base item's slot, the strongest first, then by name; extra), named for c
// (nil: a shop's goods).
func (e *RuleState) Listed(c *Character, items []*Item) []*Item {
	for _, it := range items {
		e.ItemName(c, it, false)
	}
	out := slices.Clone(items)
	slices.SortStableFunc(out, func(ia, ib *Item) int {
		if d := e.readySlot(ia.Type) - e.readySlot(ib.Type); d != 0 {
			return d
		}
		if d := e.strength(ib) - e.strength(ia); d != 0 {
			return d
		}
		return strings.Compare(ia.Name(), ib.Name())
	})
	return out
}

// strength orders items of a kind: twice the average damage of a weapon,
// twice the armour class of armour, with the magic plus.
func (e *RuleState) strength(it *Item) int {
	b := e.Base(it.Type)
	plus := int(int8(it.Plus))
	if b.givesAC() {
		return 2 * (b.AcBetter() + plus)
	}
	return int(b.Dice)*(int(b.Sides)+1) + 2*(int(int8(b.Bonus))+plus)
}

// SavedItems is what SaveCharacter writes to the .itm file: in a saved
// game the first member's items carry the party inventory (unreadied).
func (e *RuleState) SavedItems(c *Character, name string) []*Item {
	if name != "" && len(e.Party) > 0 && c == e.Party[0] {
		return e.Carried(c)
	}
	return c.Items
}
