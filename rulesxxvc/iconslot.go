package rulesxxvc

// The combat icon slots (c_combat.pas FIRST_CREATURE_ICON_SLOT, saveload.pas AddCharToParty): a party member's picture is
// loaded into the slot its orderNumber names, the monsters' from this slot on.
const (
	firstCreatureIconSlot = 8
	noIconSlot            = 255 // AddCharToParty clears orderNumber first, so the member does not count itself
)

// assignIconSlot is AddCharToParty's orderNumber: the first slot below firstCreatureIconSlot no other party member holds.
func (r *Rules) assignIconSlot(c *Character) {
	c.Order = noIconSlot
	var used [firstCreatureIconSlot]bool
	for _, id := range r.party {
		if p := r.chars[id]; p != nil && p.Order >= 0 && p.Order < firstCreatureIconSlot {
			used[p.Order] = true
		}
	}
	c.Order = 0
	for c.Order < firstCreatureIconSlot && used[c.Order] {
		c.Order++
	}
}
