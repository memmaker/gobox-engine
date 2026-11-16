package rules1e

import (
	"slices"

	"gobox/rules"
)

// The Gold Box character record's race and alignment bytes (POR's order,
// the scripts' SelRace/SelAlignment words).
var (
	RecordRace = [8]rules.Race{RaceMonster, RaceDwarf, RaceElf, RaceGnome, RaceHalfElf, RaceHalfling, RaceHalfOrc, RaceHuman}
	// RecordAlignment: LG, LN, LE, NG, N, NE, CG, CN, CE.
	RecordAlignment = [9]rules.Alignment{LawfulGood, LawfulNeutral, LawfulEvil, NeutralGood, TrueNeutral,
		NeutralEvil, ChaoticGood, ChaoticNeutral, ChaoticEvil}
)

// RecordSex is the record's gender byte (genderNameTable).
var RecordSex = [2]rules.Sex{rules.Male, rules.Female}

// RecordClassSets is each class byte's classes (the order of ClassNames, pod C_Class): the single
// classes 0..7, then the multi-class combinations.
var RecordClassSets = [len(ClassNames)]rules.ClassSet{
	ClsCleric: {ClassCleric}, ClsDruid: {ClassDruid}, ClsFighter: {ClassFighter}, ClsPaladin: {ClassPaladin},
	ClsRanger: {ClassRanger}, ClsMage: {ClassMage}, ClsThief: {ClassThief}, ClsMonk: {ClassMonk},
	clsClericFighter: {ClassCleric, ClassFighter}, clsClericFighterMage: {ClassCleric, ClassFighter, ClassMage},
	ClsClericRanger: {ClassCleric, ClassRanger}, clsClericMage: {ClassCleric, ClassMage},
	clsClericThief: {ClassCleric, ClassThief}, clsFighterMage: {ClassFighter, ClassMage},
	ClsFighterThief: {ClassFighter, ClassThief}, clsFighterMageThief: {ClassFighter, ClassMage, ClassThief},
	clsMageThief: {ClassMage, ClassThief},
}

// SaveOrder is the saving throws in the Gold Box games' order: a record's
// save bytes and a script's save number (DAMAGE) index it.
var SaveOrder = [5]rules.Save{SaveParalysis, SavePetrify, SaveRod, SaveBreath, SaveSpell}

// RaceNumbers is the race each race byte of a record of layout l names (the
// game's own numbering, which only the record codec knows: everything else
// has the typed race); "" names none. The same order as the game's race rows
// (Layout.Races).
func RaceNumbers(l rules.RecordLayout) []rules.Race {
	switch l {
	case Layout439: // Secret of the Silver Blades
		return []rules.Race{"", RaceElf, RaceHalfElf, RaceDwarf, RaceGnome, RaceHalfling, RaceHuman, RaceMonster}
	case Layout510: // Pools of Darkness
		return []rules.Race{RaceElf, RaceHalfElf, RaceDwarf, RaceGnome, RaceHalfling, RaceHuman, RaceMonster}
	}
	return RecordRace[:] // POR and COTAB
}

// raceNumber is POR's race number of s's race (the record's own byte where
// it names no player race): what the original's tests of the race byte read.
func (s Sheet) raceNumber() int8 {
	if !noPlayerRace(s.Race) {
		return int8(raceRow[s.Race])
	}
	if int(s.MonsterRace) >= len(RecordRace) {
		return int8(s.MonsterRace)
	}
	return 0
}

// SetPorRace sets the race POR's record byte n names (monster: n kept); the
// goldens and tests that have the original's race numbers use it.
func (s *Sheet) SetPorRace(n byte) { s.Race, s.MonsterRace = RaceOfRecord(Layout285, n) }

// RaceName is the race's name as the sheet shows it (DS:0F03).
func RaceName(r rules.Race) string { return RaceNames[raceRow[r]] }

// RaceOfRecord is the race of record byte b of layout l, and the byte to keep
// where it names no player race (a monster's: 0 or another number).
func RaceOfRecord(l rules.RecordLayout, b byte) (rules.Race, byte) {
	if nums := RaceNumbers(l); int(b) < len(nums) && nums[b] != "" && nums[b] != RaceMonster {
		return nums[b], 0
	}
	return RaceMonster, b
}

// recordRaceByte is the byte of a record of layout l for race (monster: the
// kept number); false where the game has no such race.
func recordRaceByte(l rules.RecordLayout, race rules.Race, monster byte) (byte, bool) {
	if noPlayerRace(race) {
		return monster, true
	}
	i := slices.Index(RaceNumbers(l), race)
	return byte(i), i >= 0
}

// noPlayerRace: the race is the monster's (or a zero Sheet's, none).
func noPlayerRace(r rules.Race) bool { return r == RaceMonster || r == "" }
