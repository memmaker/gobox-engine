package rules1e

// The import of older saves (ruling U10): everything a save written before
// a format or rule change needs to load is mended here, once, at load; saves
// are always written in the current format. This file is the whole of the
// backwards compatibility and can be cut as one piece.

// Readied values an older save may hold for the Missiles and Launcher
// slots it kept apart: the slots are back (extra), set by the item's base
// record (readySlot), so the item is simply readied.
const (
	ReadiedOldMissiles = 2
	ReadiedOldLauncher = 3
)

// ImportCharacter mends a character read from a save file.
func (e *RuleState) ImportCharacter(c *Character) {
	var armour *Item
	for _, it := range c.Items {
		if it.Readied == ReadiedOldMissiles || it.Readied == ReadiedOldLauncher { // into the slot its base record names, else back in the pack
			it.Readied = 0
			if s := e.readySlot(it.Type); s == SlotMissiles || s == SlotLauncher || e.isKind(it.Type, ItemIDArrow, ItemIDQuarrel) {
				it.Readied = 1
			}
		}
		// items saved before MendItem are mended; a readied one that only
		// now grants its affect gets it as readying would have
		had := it.Spells[2]
		e.T().MendItem(it)
		if had != it.Spells[2] && it.Readied != 0 && it.Spells[2] > 0x7F {
			e.itemAffect(it, c, 0)
		}
		// armour and bracers share the armour slot now: with both readied the
		// bracers come off
		if it.Readied == 0 || it.Readied == ReadiedOffHand || e.readySlot(it.Type) != SlotArmor {
			continue
		}
		if armour == nil {
			armour = it
			continue
		}
		off := it
		if !e.isKind(off.Type, ItemIDBracers) {
			off, armour = armour, it
		}
		off.Readied = 0
	}
}
