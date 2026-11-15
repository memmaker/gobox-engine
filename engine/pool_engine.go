package engine

import "gobox/rules"

// invItem is it as the remaster names and pictures it (the icon's key);
// its Look is left to the tooltip's Rules.
func (e *Engine) invItem(cID rules.CharID, itID rules.ItemHandle) *rules.ItemCard {
	x := e.itemCard(cID, itID)
	x.Look = nil
	return &x
}

// inventoryView builds the snapshot of the inventory screen.
func (e *Engine) inventoryView(cID rules.CharID) *Inventory {
	r := e.fact(cID, rules.InventoryFact{NoMagic: e.noMagic()}).(rules.InventoryView)
	v := &Inventory{Member: r.Member, Cap: r.Cap, Doll: r.Doll, Pool: r.Pool, Use: r.Use, Sell: r.Sell, Switch: r.Switch}
	for _, p := range e.powers(cID) {
		if p.Kind != PowerSpells || !v.Sell { // no spells in a shop
			v.Powers = append(v.Powers, p)
		}
	}
	return v
}
