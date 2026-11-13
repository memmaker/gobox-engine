package games

// Spell classes in the spell table; each has nine spell levels of
// SpellSlots.
const (
	podSpellClassCleric = iota
	podSpellClassDruid
	podSpellClassMage
	podSpellClassInnate // monster powers
	podSpellLevels      = 9
)

// spellEntry is one 16-byte record of the spell table.
type podSpellEntry [16]byte

func (s podSpellEntry) class() byte { return s[0] }

// spellTable: the record of spell n is spellTable[n].
var podSpellTable = func() (t [len(podSpellTableData)]podSpellEntry) {
	for i, r := range podSpellTableData {
		t[i] = podSpellEntry(r)
	}
	return
}()
