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

// careerBaseCombat is careerBaseCombat by career and level 0..9.
var careerBaseCombat = map[rules.Class][10]int{
	xxvc.RocketJock: {40, 40, 40, 41, 41, 42, 42, 43, 43, 44},
	xxvc.Medic:      {40, 40, 40, 40, 42, 42, 42, 44, 44, 44},
	xxvc.Warrior:    {40, 40, 41, 42, 43, 44, 45, 46, 47, 48},
	xxvc.Engineer:   {40, 40, 40, 40, 42, 42, 42, 44, 44, 44},
	xxvc.Rogue:      {40, 40, 40, 41, 41, 42, 42, 43, 43, 44},
	xxvc.Scout:      {40, 40, 41, 42, 43, 44, 45, 46, 47, 48},
}

// experienceNeeded is experienceNeeded for levels 2..9 (index level-2).
var experienceNeeded = map[rules.Class][8]int64{
	xxvc.RocketJock: {1250, 2500, 5000, 10000, 20000, 40000, 70000, 110000},
	xxvc.Medic:      {1500, 3000, 6000, 12000, 24000, 48000, 96000, 192000},
	xxvc.Warrior:    {2000, 4000, 8000, 16000, 32000, 64000, 125000, 250000},
	xxvc.Engineer:   {1250, 2500, 5000, 10000, 20000, 40000, 70000, 110000},
	xxvc.Rogue:      {1250, 2500, 5000, 10000, 20000, 40000, 70000, 110000},
	xxvc.Scout:      {2250, 4500, 9000, 18000, 36000, 75000, 150000, 300000},
}

// constitutionLevelBonus is constitutionLevelBonus by constitution 1..22 (index score-1): the bonus to a level's hit points.
var constitutionLevelBonus = [22]int{-3, -2, -2, -1, -1, -1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 5, 6, 6}

// The new character's fixed fields (chargen.pas CreateCharacterMenu).
const (
	startBaseArmorClass = 50   // baseArmorClass := 50
	startCredits        = 1000 // creditTotal := 1000
	startIconDimensions = 1    // iconDimensions := 1
	startLevel          = 2    // free training stops at the level the starting experience (experienceNeeded[career, 2]) buys
	fixedGainLevel      = 9    // chargen.pas FIXED_GAIN_LEVEL: from it a level's hit points are fixed, not rolled
	maxStartScore       = 18   // Modify keeps each ability between raceStatAdjust + 3 and + 18
	minStartScore       = 3
)

// raceMovement is baseMovement by race (CreateCharacterMenu's case; the races it offers).
var raceMovement = map[rules.Race]int{
	xxvc.Terran: 12, xxvc.Martian: 12, xxvc.Venusian: 12, xxvc.Mercurian: 12, xxvc.Tinker: 9, xxvc.DesertRunner: 15,
}

// freeAmmo is Shop.BuildAmmoItems: freeAmmoTemplates with freeAmmoQuantities, each put at the head of the item list, readied.
var freeAmmo = []struct {
	Item     rules.ItemID
	Quantity int32
}{{7, 100}, {20, 0}}
