package rules1e

// The shop, ported from ovr07. ECL COMBAT with the shop flag (game +0x6D8)
// set runs Shop instead of a fight; the goods are the treasure list.

// BuyItem = ovr07:0239 (item, var overloaded): a copy goes to the
// selected character unless it is too much to carry.
func (e *RuleState) BuyItem(it *Item) (overloaded bool) {
	c := e.Acting()
	if e.Overloaded(it, c) {
		e.notice("No room in the party items") // extra: over the inventory screen (the original: OverLoaded)
		return true
	}
	e.AddItemCopy(c, it)
	e.ApplyItems(c)
	return false
}

// Identifiable is member i's items Id looks at (extra: the party
// inventory with the first member).
func (e *RuleState) Identifiable(i int) []*Item {
	if i == 0 {
		return e.Carried(e.Party[0])
	}
	return e.Party[i].Items
}

// CanUseItem (extra): the item's class mask lets c ready it (the Wrong
// Class test of readyItem).
func (e *RuleState) CanUseItem(c *Character, it *Item) bool {
	return e.Base(it.Type).Classes&c.ClassMask != 0
}
