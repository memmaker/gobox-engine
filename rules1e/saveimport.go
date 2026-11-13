package rules1e

import "gobox/core/dnd"

// The import of older saves (ruling U10): everything a save written before
// a format or rule change needs to load is mended here, once, at load; saves
// are always written in the current format. This file is the whole of the
// backwards compatibility and can be cut as one piece.

// Readied values an older save may hold for slots that no longer exist.
const (
	ReadiedOldMissiles = 2 // the Missiles slot
	ReadiedOldLauncher = 3 // the Launcher slot
)

// ImportCharacter mends a character read from a save file.
func (e *RuleState) ImportCharacter(c *Character) {
	var armour *Item
	for _, it := range c.Items {
		if it.Readied == ReadiedOldMissiles || it.Readied == ReadiedOldLauncher { // back in the pack
			it.Readied = 0
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
		if !e.isKind(off.Type, dnd.ItemBracers) {
			off, armour = armour, it
		}
		off.Readied = 0
	}
}
