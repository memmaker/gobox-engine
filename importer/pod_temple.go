package importer

// templeServiceSpells = DS:08CB[1..11]: the spells the Heal list names.
var templeServiceSpells = [...]byte{37, 39, 3, 58, 71, 36, 67, spellRaiseDead, 43, 113} // DS:08CB's 11th, Restoration, is gone (Felix 2026-11-12)

// templeServices: the Heal list (O04C:062E), the spells' names.
var podTempleServices = func() (names [len(templeServiceSpells)]string) {
	for i, sp := range templeServiceSpells {
		names[i] = podSpellNames[sp]
	}
	return
}()
