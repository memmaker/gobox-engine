// Package xxvc is the Buck Rogers concepts: what BR1, BR2 and the XXVc ruleset
// name, from buckroger-re and buckrogers2-re (gdconsts.pas, c_skill.pas,
// c_sys.pas; the rosters are the same in both games). Monsters, items and ships
// differ by game: they are Tables, written by the importer.
package xxvc

import "gobox/core"

const Tech core.Ability = "tech" // the seventh score (eclbr.OffStats)

// The races (types.pas TRace, raceNames), RaceMonster left out.
const (
	Terran       core.Race = "terran"
	Martian      core.Race = "martian"
	Venusian     core.Race = "venusian"
	Mercurian    core.Race = "mercurian"
	Tinker       core.Race = "tinker"
	DesertRunner core.Race = "desert-runner"
	Lunarian     core.Race = "lunarian"
	Lowlander    core.Race = "lowlander"
)

// Races in the original's order (RaceTerran..RaceLowlander).
var Races = []core.Race{Terran, Martian, Venusian, Mercurian, Tinker, DesertRunner, Lunarian, Lowlander}

// The careers (careerNames). No race may take Scout (raceCareers).
const (
	RocketJock core.Class = "rocket-jock"
	Medic      core.Class = "medic"
	Warrior    core.Class = "warrior"
	Engineer   core.Class = "engineer"
	Rogue      core.Class = "rogue"
	Scout      core.Class = "scout"
)

// Careers in the original's order (CareerRocketJock..CareerScout).
var Careers = []core.Class{RocketJock, Medic, Warrior, Engineer, Rogue, Scout}

const BattleTactics core.Skill = "battle-tactics" // BR skill 31 (combat.go tactics)

// Skills are the 84 skills by number: Skills[n-1] is skill n (skillDefinitions,
// TCharacter.skillLevels' index).
var Skills = []core.Skill{
	"repair-electrical", "repair-mechanical", "repair-nuclear-engine", "repair-life-support",
	"repair-rocket-hull", "jury-rig", "bypass-security", "open-lock", "commo-operation",
	"sensor-operation", "demolitions", "cook", "first-aid", "repair-weapon",
	"repair-computer", "life-suspension-tech", "treat-light-wounds", "treat-serious-wounds",
	"treat-critical-wounds", "treat-poisoning", "treat-stun-paralysis", "treat-disease", "diagnose",
	"general-knowledge", "ship-lore", "composition", "mathematics", "physics", "geology",
	"planetology", BattleTactics, "chemistry", "biology", "design-engineering", "history",
	"astronomy", "botany", "literature", "economics", "metallurgy", "speak-read-language", "law",
	"programming", "cryptography", "library-search", "astrogation", "navigation", "mimic",
	"memorize", "disguise", "pilot-rocket", "pilot-fixed-wing", "drive-ground-car",
	"pilot-rotor-wing", "drive-motorcycle", "use-jetpack", "pilot-ship-sub",
	"drive-heavy-ground-vehicle", "drive-jetcar", "animal-riding", "hide-in-shadows",
	"move-silently", "pick-pockets", "acrobatics", "climb", "swimming", "maneuver-in-0g",
	"play-instrument", "paint-draw", "hypnosis", "act", "intimidate", "leadership",
	"befriend-animal", "animal-training", "fast-talk-convince", "singing", "distract", "etiquette",
	"tracking", "shadowing", "read-lips", "notice", "planetary-survival",
}

// Saves are the eight saving throws by index (TCharacter.saves, an effect's
// saveType). The original names none of them.
var Saves = []core.Save{"save-0", "save-1", "save-2", "save-3", "save-4", "save-5", "save-6", "save-7"}

const Credits core.Coin = "credits"

// A ship's systems, each with a now and a max (c_sys.pas, shipSystemNames).
const (
	Hull        core.ShipSystem = "hull"
	Sensors     core.ShipSystem = "sensors"
	Control     core.ShipSystem = "control"
	LifeSupport core.ShipSystem = "life-support"
	Fuel        core.ShipSystem = "fuel"
	Engines     core.ShipSystem = "engines"
)

// ShipSystems in the original's order (TShip.systemStatus' index).
var ShipSystems = []core.ShipSystem{Hull, Sensors, Control, LifeSupport, Fuel, Engines}

// Tables are a game's numbers that differ between BR1 and BR2 (START.EXE's
// data segment), written by the importer and passed at Begin as
// core.Campaign.Tables.
type Tables struct {
	ItemTypes []ItemType // by TItemTemplate.itemType
	ItemWords []string   // the words items' names are made of, by number (1..; 0 none)
	Items     map[core.ItemID]ItemTemplate
	Monsters  map[core.MonsterID]Monster
}

// ItemTemplate is one row of itemTemplates (types.pas TItemTemplate): what a
// new item copies. Its ItemID is the importer's, stable by assignment.
type ItemTemplate struct {
	Type        int32    // into ItemTypes
	Words       [3]int32 // into ItemWords, 0 none
	Plus        int32
	SaveBonus   int32
	NotTreasure bool // C_Item.NOT_TREASURE: never loot
	NotJoinable bool // C_Item.NOT_JOINABLE
	Weight      int32
	Quantity    int32
	Value       int32
	Special     [3]int32 // special 2 a projectile kind, fired while special 3 is non-zero
}

// Monster is one monster record (a TCharacter that EclMenu.InitCreature loads):
// its rule fields, typed. Its MonsterID is the importer's.
type Monster struct {
	Name        string
	Race        core.Race
	Career      core.Class
	Level       int32
	Scores      map[core.Ability]int32
	HP          int32
	BaseCombat  int32 // baseCombat, attackBonus starts from it
	AttackBonus int32
	ArmorClass  [2]int32     // front, rear
	Damage      [2]core.Dice // per attack phase (currentRolls, currentDice, currentModifiers)
	BaseScale   int32        // phase 1's attacks per round times two, unarmed (baseScale)
	Scale       int32        // attacks per round times two (scale)
	Movement    int32
	Skills      map[core.Skill]int32
	Saves       [8]int32
	Hostile     bool // C_Side.HOSTILE; else on the team's side
	XPAward     int64
	Credits     core.Copper // creditTotal
	Items       []core.ItemRef
}

func (Tables) RuleTables() {}

// EquipSlot is where an item is readied (c_equip.pas).
type EquipSlot int32

const (
	SlotWeapon          EquipSlot = 0
	SlotArmor           EquipSlot = 1
	SlotJetpack         EquipSlot = 2
	SlotGrenadeLauncher EquipSlot = 6
	SlotECM             EquipSlot = 7
)

// WeaponClass is a weapon's kind, which effects and armor test (c_wclass.pas).
type WeaponClass int32

const (
	Melee          WeaponClass = 0
	Slug           WeaponClass = 1
	Rocket         WeaponClass = 2
	Laser          WeaponClass = 3
	Microwave      WeaponClass = 4
	Heat           WeaponClass = 5 // the Heat Gun
	Sonic          WeaponClass = 6 // the Sonic Stunner
	RocketLauncher WeaponClass = 7 // this and the next two are area weapons
	Grenade        WeaponClass = 8
	Plasma         WeaponClass = 9
)

// ItemType is one row of itemTypes (types.pas TItemType). The original keeps
// an armor's class plus 50 in weaponClass; here it is ArmorClass.
type ItemType struct {
	Slot           EquipSlot
	AttacksDoubled int32 // attacks per round times two
	WeaponClass    WeaponClass
	ArmorClass     int32 // armor only
	Heavy          bool  // readying it in combat ends the turn
	Damage         core.Dice
	Range          int32 // cells; above 1 is ranged
	Consumable     bool
}
