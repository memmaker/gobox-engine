package rulesxxvc

import (
	"gobox/rules"
	"gobox/rules/xxvc"
)

// The race and career rules a character is made and trained by (gdconsts.pas, "The race and career rules"). The
// per-ability rows are in brStatOrder (STR DEX CON INT WIS CHA TCH); each table is keyed by race or career id.

// pickRaces are the races CreateCharacterMenu offers (RaceTerran .. RaceDesertRunner), in its order.
var pickRaces = []rules.Race{xxvc.Terran, xxvc.Martian, xxvc.Venusian, xxvc.Mercurian, xxvc.Tinker, xxvc.DesertRunner}

// raceCareers is raceCareers: the careers each race may take, the first `count` of its row (no race takes Scout).
var raceCareers = map[rules.Race][]rules.Class{
	xxvc.Terran:       {xxvc.RocketJock, xxvc.Medic, xxvc.Warrior, xxvc.Engineer, xxvc.Rogue},
	xxvc.Martian:      {xxvc.RocketJock, xxvc.Medic, xxvc.Warrior, xxvc.Engineer, xxvc.Rogue},
	xxvc.Venusian:     {xxvc.RocketJock, xxvc.Medic, xxvc.Warrior, xxvc.Engineer, xxvc.Rogue},
	xxvc.Mercurian:    {xxvc.RocketJock, xxvc.Medic, xxvc.Warrior, xxvc.Engineer, xxvc.Rogue},
	xxvc.Tinker:       {xxvc.Medic, xxvc.Engineer},
	xxvc.DesertRunner: {xxvc.RocketJock, xxvc.Warrior, xxvc.Engineer},
	xxvc.Lunarian:     {xxvc.RocketJock, xxvc.Medic, xxvc.Warrior, xxvc.Engineer, xxvc.Rogue},
	xxvc.Lowlander:    {xxvc.Warrior, xxvc.Engineer},
}

// raceStatAdjust is raceStatAdjust: added to each rolled ability.
var raceStatAdjust = map[rules.Race][7]int{
	xxvc.Terran:       {0, 0, 1, 0, 1, 0, 0},
	xxvc.Martian:      {-1, 1, -1, 0, -1, 1, 0},
	xxvc.Venusian:     {0, -1, 1, 0, 1, -1, 0},
	xxvc.Mercurian:    {-1, 1, 1, 0, 0, 0, 0},
	xxvc.Tinker:       {-2, 3, -2, 0, 0, 0, 3},
	xxvc.DesertRunner: {2, 2, 1, 0, 0, -1, 0},
	xxvc.Lunarian:     {-2, 1, -1, 2, 0, 0, 0},
	xxvc.Lowlander:    {3, 0, 1, 0, 0, -3, 1},
}

// careerStatMinimum is careerStatMinimum: each career's lowest ability.
var careerStatMinimum = map[rules.Class][7]int{
	xxvc.RocketJock: {1, 13, 1, 11, 1, 12, 1},
	xxvc.Medic:      {1, 12, 1, 12, 12, 1, 1},
	xxvc.Warrior:    {10, 8, 10, 8, 1, 1, 1},
	xxvc.Engineer:   {10, 1, 12, 8, 1, 1, 13},
	xxvc.Rogue:      {1, 13, 1, 8, 9, 13, 1},
	xxvc.Scout:      {1, 1, 8, 8, 9, 8, 1},
}

// baseSavingThrows is baseSavingThrows; raceSaveAdjust is subtracted from it (InitializeSaves).
var baseSavingThrows = [8]int{15, 14, 14, 13, 13, 16, 13, 13}

var raceSaveAdjust = map[rules.Race][8]int{
	xxvc.Terran:       {1, 0, 0, 0, 0, 0, 0, 0},
	xxvc.Martian:      {-1, 0, 0, 0, 1, 1, -1, 1},
	xxvc.Venusian:     {1, 3, 0, 1, 1, -2, 3, -2},
	xxvc.Mercurian:    {1, 0, 0, 0, 3, 1, 4, -3},
	xxvc.Tinker:       {0, -1, -3, -2, 0, 0, 0, 0},
	xxvc.DesertRunner: {2, 0, 1, 1, 2, 2, -3, 4},
	xxvc.Lunarian:     {0, -1, 0, 0, 2, 2, -1, -1},
	xxvc.Lowlander:    {2, 4, 2, 3, 1, -4, 6, -6},
}

// constitutionLevelBonus is constitutionLevelBonus by constitution 1..22 (index score-1): the bonus to a level's hit points.
var constitutionLevelBonus = [22]int{-3, -2, -2, -1, -1, -1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 5, 6, 6}

// The new character's fixed fields (chargen.pas CreateCharacterMenu).
const (
	startBaseArmorClass       = 50   // baseArmorClass := 50
	startCredits              = 1000 // creditTotal := 1000
	startIconDimensions       = 1    // iconDimensions := 1
	lastSpecializableTemplate = 15   // viewchar.pas LAST_SPECIALIZABLE_TEMPLATE: the menu offers itemTemplates 1..15
	ageBase                   = 18   // age := RollDice(1, 6) + 18
	ageDie                    = 6
	maxAge                    = 255 // age is a Byte
	maxStartScore             = 18  // Modify keeps each ability between raceStatAdjust + 3 and + 18
	minStartScore             = 3
)

// raceMovement is baseMovement by race (CreateCharacterMenu's case; the races it offers).
var raceMovement = map[rules.Race]int{
	xxvc.Terran: 12, xxvc.Martian: 12, xxvc.Venusian: 12, xxvc.Mercurian: 12, xxvc.Tinker: 9, xxvc.DesertRunner: 15,
}

// The starting icons (chargen.pas SetDefaultIcon, gdconsts.pas characterIcons): a Tinker's, one of four at random for
// a Desert Runner, else by gender and career (no Scout column: no race takes Scout).
const tinkerIcon = 17

var desertRunnerIcons = [4]int{16, 23, 24, 31} // RollDice(1, 4) picks one

var characterIcons = map[rules.Sex]map[rules.Class]int{
	rules.Male:   {xxvc.RocketJock: 32, xxvc.Medic: 0, xxvc.Warrior: 3, xxvc.Engineer: 7, xxvc.Rogue: 21},
	rules.Female: {xxvc.RocketJock: 33, xxvc.Medic: 1, xxvc.Warrior: 6, xxvc.Engineer: 15, xxvc.Rogue: 25},
}

// defaultIcon is SetDefaultIcon's icon for a new character.
func defaultIcon(race rules.Race, sex rules.Sex, career rules.Class, rng *rules.RNG) int {
	switch race {
	case xxvc.Tinker:
		return tinkerIcon
	case xxvc.DesertRunner:
		return desertRunnerIcons[rules.RNGDice(rng, 1, len(desertRunnerIcons))-1]
	}
	return characterIcons[sex][career]
}

// freeAmmo is Shop.BuildAmmoItems: freeAmmoTemplates with freeAmmoQuantities, each put at the head of the item list, readied.
var freeAmmo = []struct {
	Item     rules.ItemID
	Quantity int32
}{{7, 100}, {20, 0}}
