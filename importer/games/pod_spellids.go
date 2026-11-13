package games

// Spell numbers (the spell table, the spellbook and memorised lists).
const (
	spellHeal                   = 36
	spellHarm                   = 56
	spellPotionOfSpeed          = 57
	spellPotionOfFrost          = 59
	spellJavelinOfLightning     = 60
	spellWandOfParalyzation     = 61
	spellPotionOfHealing        = 62
	spellElixirOfYouth          = 63
	spellNecklaceOfMissiles     = 64
	spellWandOfMagicMissiles    = 65
	spellRemoveCurseMU          = 89
	spellBarkskin               = 90
	spellScrollOfProtection     = 95
	spellCharmPersonOrMammal    = 96
	spellPotionOfInvisibility   = 97
	spellCureLightWoundsDruid   = 98
	spellPotionOfExtraHealing   = 99
	spellBladeBarrier           = 101
	spellDestruction            = 104
	spellResurrection           = 105
	spellCureDiseaseDruid       = 106
	spellNeutralizePoisonDruid  = 107
	spellHoldAnimal             = 108
	spellProtectFromFire        = 109
	spellDeathSpell             = 110
	spellDisintegrate           = 111
	spellGlobeOfInvulnerability = 112
	spellStoneToFlesh           = 113
	spellFleshToStone           = 114
	spellDelayedBlastFireball   = 115
	spellMassInvisibility       = 116
	spellPowerWordStun          = 117
	spellFireTouch              = 118
	spellIronSkin               = 119
	spellMassCharm              = 120
	spellOttosIrresistibleDance = 121
	spellMindBlank              = 122
	spellPowerWordBlind         = 123
	spellMeteorSwarm            = 124
	spellPowerWordKill          = 125
	spellMonsterSummoning       = 126
	spellCount                  = 126 // the highest spell number
)

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
