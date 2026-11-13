package importpod

// Spell numbers (the spell table, the spellbook and memorised lists).
const (
	spellBless                         = 1
	spellCurse                         = 2
	spellCureLightWounds               = 3
	spellCauseLightWounds              = 4
	spellDetectMagicCleric             = 5
	spellProtectFromEvilCleric         = 6
	spellProtectFromGoodCleric         = 7
	spellResistCold                    = 8
	spellBurningHands                  = 9
	spellCharmPerson                   = 10
	spellDetectMagicMU                 = 11
	spellEnlarge                       = 12
	spellReduce                        = 13
	spellFriends                       = 14
	spellMagicMissile                  = 15
	spellProtectFromEvilMU             = 16
	spellProtectFromGoodMU             = 17
	spellReadMagic                     = 18
	spellShield                        = 19
	spellShockingGrasp                 = 20
	spellSleep                         = 21
	spellFindTraps                     = 22
	spellHoldPersonCleric              = 23
	spellResistFire                    = 24
	spellSilence15Radius               = 25
	spellSlowPoison                    = 26
	spellSnakeCharm                    = 27
	spellSpiritualHammer               = 28
	spellDetectInvisibility            = 29
	spellInvisibility                  = 30
	spellKnock                         = 31
	spellMirrorImage                   = 32
	spellRayOfEnfeeblement             = 33
	spellStinkingCloud                 = 34
	spellStrength                      = 35
	spellHeal                          = 36
	spellCureBlindness                 = 37
	spellCauseBlindness                = 38
	spellCureDisease                   = 39
	spellCauseDisease                  = 40
	spellDispelMagicCleric             = 41
	spellPrayer                        = 42
	spellRemoveCurse                   = 43
	spellBestowCurseCleric             = 44
	spellBlink                         = 45
	spellDispelMagicMU                 = 46
	spellFireball                      = 47
	spellHaste                         = 48
	spellHoldPersonMU                  = 49
	spellInvisibility10Radius          = 50
	spellLightningBolt                 = 51
	spellProtectFromEvil10Radius       = 52
	spellProtectFromGood10Radius       = 53
	spellProtectFromNormalMissiles     = 54
	spellSlow                          = 55
	spellHarm                          = 56
	spellPotionOfSpeed                 = 57
	spellCureSeriousWounds             = 58
	spellPotionOfFrost                 = 59
	spellJavelinOfLightning            = 60
	spellWandOfParalyzation            = 61
	spellPotionOfHealing               = 62
	spellElixirOfYouth                 = 63
	spellNecklaceOfMissiles            = 64
	spellWandOfMagicMissiles           = 65
	spellCauseSeriousWounds            = 66
	spellNeutralizePoison              = 67
	spellPoison                        = 68
	spellProtectFromEvil10RadiusCleric = 69
	spellSticksToSnakes                = 70
	spellCureCriticalWounds            = 71
	spellCauseCriticalWounds           = 72
	spellDispelEvil                    = 73
	spellFlameStrike                   = 74
	spellRaiseDead                     = 75
	spellSlayLiving                    = 76
	spellDetectMagicDruid              = 77
	spellEntangle                      = 78
	spellFaerieFire                    = 79
	spellInvisibilityToAnimals         = 80
	spellCharmMonsters                 = 81
	spellConfusion                     = 82
	spellDimensionDoor                 = 83
	spellFear                          = 84
	spellFireShield                    = 85
	spellFumble                        = 86
	spellIceStorm                      = 87
	spellMinorGlobeOfInvulnerability   = 88
	spellRemoveCurseMU                 = 89
	spellBarkskin                      = 90
	spellCloudKill                     = 91
	spellConeOfCold                    = 92
	spellFeeblemind                    = 93
	spellHoldMonsters                  = 94
	spellScrollOfProtection            = 95
	spellCharmPersonOrMammal           = 96
	spellPotionOfInvisibility          = 97
	spellCureLightWoundsDruid          = 98
	spellPotionOfExtraHealing          = 99
	spellBestowCurseMU                 = 100
	spellBladeBarrier                  = 101
	spellDestruction                   = 104
	spellResurrection                  = 105
	spellCureDiseaseDruid              = 106
	spellNeutralizePoisonDruid         = 107
	spellHoldAnimal                    = 108
	spellProtectFromFire               = 109
	spellDeathSpell                    = 110
	spellDisintegrate                  = 111
	spellGlobeOfInvulnerability        = 112
	spellStoneToFlesh                  = 113
	spellFleshToStone                  = 114
	spellDelayedBlastFireball          = 115
	spellMassInvisibility              = 116
	spellPowerWordStun                 = 117
	spellFireTouch                     = 118
	spellIronSkin                      = 119
	spellMassCharm                     = 120
	spellOttosIrresistibleDance        = 121
	spellMindBlank                     = 122
	spellPowerWordBlind                = 123
	spellMeteorSwarm                   = 124
	spellPowerWordKill                 = 125
	spellMonsterSummoning              = 126
	spellCount                         = 126 // the highest spell number
)

// Spell classes in the spell table; each has nine spell levels of
// SpellSlots.
const (
	spellClassCleric = iota
	spellClassDruid
	spellClassMage
	spellClassInnate // monster powers
	spellLevels      = 9
)

// spellEntry is one 16-byte record of the spell table.
type spellEntry [16]byte

func (s spellEntry) class() byte { return s[0] }

// spellTable: the record of spell n is spellTable[n].
var spellTable = func() (t [len(spellTableData)]spellEntry) {
	for i, r := range spellTableData {
		t[i] = spellEntry(r)
	}
	return
}()
