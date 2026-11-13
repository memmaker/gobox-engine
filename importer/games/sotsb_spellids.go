package games

import "strconv"

// Spell numbers (the spell table, the spellbook and memorised lists).
const (
	spellSpell101 = 101
	spellSpell102 = 102
	spellSpell103 = 103
	spellSpell104 = 104
	spellSpell105 = 105
	spellSpell106 = 106
	spellSpell107 = 107
	spellSpell108 = 108
	spellSpell109 = 109
)

// Spell classes in the spell table; each has seven spell levels of
// SpellSlots (the third group is unused).
const (
	sotsbSpellClassCleric = iota
	sotsbSpellClassDruid
	_
	sotsbSpellClassMage
	sotsbSpellClassInnate // monster powers
	sotsbSpellLevels      = 7
)

// Spell table record fields (spellRec).
const (
	spellFieldClass         = 0
	spellFieldLevel         = 1
	spellFieldRange         = 2
	spellFieldRangePerLevel = 3
	spellFieldAmount        = 4
	spellFieldAmountPerLvl  = 5
	spellFieldLasts         = 6  // the range is at least 1
	spellFieldTargets       = 7  // 0: combat only; how the targets are chosen
	spellFieldSaveMode      = 8  // what a saving throw does
	spellFieldSaveType      = 9  // the saving throw kind
	spellFieldAffect        = 10 // the affect the spell lays
	spellEndsTurn           = 11 // casting it ends the turn
)

// spellEntry is one 16-byte record of the spell table.
type sotsbSpellEntry [16]byte

// spellTable: the record of spell n is spellTable[n].
var sotsbSpellTable = func() (t [len(sotsbSpellTableData)]sotsbSpellEntry) {
	for i, r := range sotsbSpellTableData {
		t[i] = sotsbSpellEntry(r)
	}
	return
}()

// tableID is a spell or item-word number from the data, any width.
type tableID interface{ ~int | ~byte | ~uint16 }

// spellName is spell id's name; past the table (a data byte the original
// reads beyond its table, bug class 9) it is "spell N", never a panic.
func spellName[T tableID](id T) string {
	if int(id) < len(sotsbSpellNames) {
		return sotsbSpellNames[id]
	}
	return "spell " + strconv.Itoa(int(id))
}

// spellRecord is spell id's record; past the table, an empty one.
func spellRecord[T tableID](id T) sotsbSpellEntry {
	if int(id) < len(sotsbSpellTable) {
		return sotsbSpellTable[id]
	}
	return sotsbSpellEntry{}
}
