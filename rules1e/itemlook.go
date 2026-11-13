package rules1e

import "gobox/core"

// Item Look (extra): the shop's Look shows what an item does under the
// game's rules, read from its base record (DS:5582 + 16*type) and the
// affect handlers that test for it.

// DamageAvg is the item's average damage (against small foes) with its
// known magic plus, for the Look's comparison; 0 for nil.
func (e *RuleState) DamageAvg(it *Item) float64 {
	if it == nil {
		return 0
	}
	b := e.Base(it.Type)
	return float64(b.Dice)*float64(b.Sides+1)/2 + float64(int8(b.Bonus)) + e.KnownPlus(it)
}

// LargeAvg is damageAvg against large foes: the large dice, or the normal
// ones when the base record has none.
func (e *RuleState) LargeAvg(it *Item) float64 {
	if it == nil {
		return 0
	}
	b := e.Base(it.Type)
	if b.LargeDice == 0 {
		return e.DamageAvg(it)
	}
	return float64(b.LargeDice)*float64(b.LargeSides+1)/2 + float64(int8(b.LargeBonus)) + e.KnownPlus(it)
}

// Protection is the armour class points the item is worth with its known
// magic plus (armour's and bracers' over AC 10), for the Look's comparison; 0 for nil.
func (e *RuleState) Protection(it *Item) float64 {
	if it == nil {
		return 0
	}
	return float64(e.Base(it.Type).AcBetter()) + e.KnownPlus(it)
}

// KnownPlus is the item's magic plus when identified, else 0.
func (e *RuleState) KnownPlus(it *Item) float64 {
	if it.Hidden != 0 {
		return 0
	}
	return float64(int8(it.Plus))
}

// ReadiedLike (extra) is the item c has readied in the doll slot it would
// take, for the Look's comparison: nil for a readied item or one without a slot.
// A weapon thrown or used in melee (hammer, spear) is set against the weapon
// in hand, not the Missiles slot's arrows.
func (e *RuleState) ReadiedLike(c *Character, it *Item) *Item {
	s := e.DollSlot(c, it)
	if it.Readied != 0 || s < 0 {
		return nil
	}
	if s == core.DollArrows && e.Base(it.Type).Usage&0x14 == 0x14 {
		s = core.DollWeapon
	}
	for _, o := range c.Items {
		if o.Readied != 0 && o != it && e.DollSlot(c, o) == s {
			return o
		}
	}
	return nil
}
