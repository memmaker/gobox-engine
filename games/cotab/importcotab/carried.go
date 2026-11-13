package importcotab

import (
	"embed"
	eclscript "gobox/ecl/script"
	"gobox/game"
	"gobox/rules1e"
	"io/fs"
	"slices"
)

// RecordLayout is the layout of the game's character records (its adventure's manifest names it).
const RecordLayout = rules1e.Layout422

// scriptFS is the scripts' data (gobox ecl/script Library.Data): the
// layout, made from cotab-re's Pascal (scripts/layout/global.txt).
//
//go:embed scripts
var scriptFS embed.FS

func scriptData() fs.FS {
	d, err := fs.Sub(scriptFS, "scripts")
	if err != nil {
		panic(err) // a constant, valid path
	}
	return d
}

// ScriptLayout is the game's script fields (scripts/layout/global.txt).
func ScriptLayout() (*eclscript.Layout, error) { return eclscript.ReadLayout(scriptData(), 0) }

// OriginalSave is savgamX.dat as above (unit O0A1): the area byte, the area, area2 and aux
// records at their script addresses, the script code, the position and states, the wall slots.
var OriginalSave = rules1e.SaveLayout{
	Lead: rules1e.LeadAreaKind,
	Records: []rules1e.SaveRecord{
		{Base: scriptAreaFirst, Cells: scriptAreaLast - scriptAreaFirst + 1, CellBytes: 2},
		{Base: scriptArea2First, Cells: scriptArea2Last - scriptArea2First + 1, CellBytes: 2},
		{Base: scriptAuxFirst, Cells: scriptAuxLast - scriptAuxFirst + 1, CellBytes: 2},
	},
	CodeBase:  scriptCodeFirst,
	CodeSize:  scriptCodeSize,
	CodeSaved: true,
	WallBytes: 3 * 4,
}

// DungeonTiles is how many tiles DungCom puts at the start of the set.
const DungeonTiles = 0x18 + 1

// Camp's rests (CampRest), as por numbers them (present.Data).
const (
	RestFix   = iota // Fix: the clerics heal the party
	RestWait         // Rest: the original's rest menu (Days, Hours ..., the Ask)
	RestUntil        // until day or night: not the original's, refused
)

// charRecord is the game's character record, the shared 422-byte layout
// (gobox/char, field for field: CHRDATA/SAVGAM and the monster records).
type charRecord = rules1e.Record422

// Rubble on the dungeon floor.
const (
	cellRubble       = 0x40 // the map cell's flag for rubble in the room
	groundRubble     = 26
	groundRubbleEdge = 27
)

// Wilderness terrain flags (wildTerrain, by area byte 0x342).
const (
	terrainSparse   = 2   // fewer trees and rocks
	terrainThin     = 4   // a few fewer
	terrainForest   = 8   // more trees
	terrainHills    = 16  // a cliff now and then
	terrainMountain = 32  // a cliff more often
	terrainRough    = 64  // more of everything
	terrainBare     = 128 // no trees, much less of anything
)

// Combat ground values of the wilderness.
const (
	groundGrass        = 23 // the whole map to start with
	tileOpenGround     = 22 // groundInfo tile of open ground: things may be put there
	groundCliffLeft    = 59 // +1d2
	groundCliffRight   = 61 // +1d2
	groundCliffFoot    = 64
	groundCliffFoot2   = 65
	groundTreeTop      = 31 // +1d5, above groundTreeTrunk
	groundTreeTrunk    = 36 // +1d5
	groundShrub        = 41 // +1d2
	groundBoulder      = 57 // +1d2
	groundBush         = 47 // +1d2
	groundRocks        = 43 // +1d4
	groundStones       = 54 // +1d3
	groundFlowers      = 49 // +1d4
	areaByteTerrain    = 0x342
	wildMapW, wildMapH = 50, 25
)

// A new character's fixed starting values.
const (
	startBaseAC        = 50
	startTHAC0         = 40
	startIconID        = 10
	startExperience    = 25000 // one class
	startExperience2   = 12500 // two classes
	startExperience3   = 8333  // three classes
	startPlatinum      = 300
	goldPerPlatinum    = 5
	startBaseMove      = 12
	exceptionalStrDice = 100
	strengthMaximum    = 18
	multiClericWisdom  = 13
	maxNameLength      = 15
)

// The ECL script machine: operands, variables, strings, comparisons, calls
// (unit O04C of the reconstruction). Script addresses are 16-bit:
//
//	19200..20223  area record words      (p4FB2)
//	31232..31743  aux record words       (p4FBA)
//	31744..32767  area2 record words     (p4FB6; some are character values)
//	32768..40447  the script code, bytes (p4FBE)
//	others        special values (direction, position, ...)
const (
	scriptAreaFirst  = 19200
	scriptAreaLast   = 20223
	scriptAuxFirst   = 31232
	scriptAuxLast    = 31743
	scriptArea2First = 31744
	scriptArea2Last  = 32767
	scriptCodeFirst  = 32768
	scriptCodeLast   = 40447
	scriptCodeSize   = scriptCodeLast - scriptCodeFirst + 1
	specialBase      = 49000 // writes above it: position and direction
	specialReadBase  = 49227 // reads above it: position, direction, ...
)

// ImportScriptsTo decompiles the game's ECL*.DAX in dos into an adventure
// folder's scripts dir, with the script data beside them (the importer's
// step; eclscript.ImportDAXFolder).
func ImportScriptsTo(dos, dir string) error {
	return eclscript.ImportDAXFolder(eclscript.V2, dos, scriptData(), dir, scriptFixes)
}

// Item and Affect are the raw list nodes (next pointers live in the slices).
// Item is the 0x3F-byte DOS item record, in its file layout (encoding/binary).
type Item struct {
	PName   [0x29]byte // 0x00 shown name, Pascal string (built by ItemName)
	Pad29   byte       // 0x29
	_       [4]byte    // 0x2A next pointer
	Type    byte       // 0x2E base item type (index into the DS:5582 records)
	Words   [3]byte    // 0x2F name word ids into DS:10A7
	Plus    byte       // 0x32 signed
	Plus2   byte       // 0x33
	Readied byte       // 0x34
	Hidden  byte       // 0x35 bit 3-i: word i not known yet
	Cursed  byte       // 0x36
	Weight  uint16     // 0x37
	Count   byte       // 0x39
	Price   uint16     // 0x3A
	Spells  [3]byte    // 0x3C a scroll's spells, +0x80 = being scribed
}

// affectRecord is the 9-byte DOS affect node in its file layout.
type affectRecord struct {
	Type byte   // affect number, index into the ovr13 handler table
	Time uint16 // remaining duration, 0 = permanent
	Data byte
	Call byte    // non-zero: the handler runs when the affect is removed
	_    [4]byte // next pointer
}

// The adventure folder's parts (gobox import cotab writes them).
const (
	scriptsDir = "scripts"
	artDir     = "art"
	dataDir    = "data"
	soundDir   = "sound"
	saveDir    = "saves"
)

// itemTypes is the base item table's records (DS:5582, 0x810 bytes).
const itemTypes = 0x81

// quickParty: a fighter and a paladin in front, a cleric, a fighter/mage,
// a magic-user and a thief for locks and traps; nobody evil (the paladin).
var quickParty = []game.QuickPick{
	{Race: "Dwarf", Gender: "Male", Class: "Fighter", Align: "Lawful Good", Need: map[string]int{"str": 18, "str00": 50, "con": 17}, Name: "Thorgrim", Head: "head_050", Body: "body_027"},
	{Race: "Human", Gender: "Male", Class: "Paladin", Align: "Lawful Good", Need: map[string]int{"str": 17, "wis": 14, "con": 15}, Name: "Aldric", Head: "head_049", Body: "body_020"},
	{Race: "Human", Gender: "Female", Class: "Cleric", Align: "Lawful Good", Need: map[string]int{"wis": 17, "con": 15, "str": 15}, Name: "Mirelle", Head: "head_044", Body: "body_023"},
	{Race: "Elf", Gender: "Female", Class: "Fighter/Magic-User", Align: "Chaotic Good", Need: map[string]int{"str": 16, "int": 17, "dex": 16}, Name: "Sylvaine", Head: "head_048", Body: "body_024"},
	{Race: "Human", Gender: "Male", Class: "Magic-User", Align: "Neutral Good", Need: map[string]int{"int": 18, "dex": 16, "con": 14}, Name: "Corwin", Head: "head_047", Body: "body_025"},
	{Race: "Halfling", Gender: "Male", Class: "Thief", Align: "Neutral Good", Need: map[string]int{"dex": 18, "con": 15}, Name: "Pip", Head: "head_042", Body: "body_021"},
}

// The sound numbers the game passes to PlaySound (the table at DS:2612); the
// driver itself is the frontend's (package audio).
const (
	sndStop       = 255 // stop the song
	sndMute       = 0   // stop and mute
	sndUnmute     = 1
	sndSpell      = 2 // a spell flies (the spells but 47 and 51)
	sndHit        = 3 // a weapon hits
	sndMagicHit   = 4 // a magical weapon hits
	sndFall       = 5 // a combatant falls
	sndMissile    = 6 // a missile flies
	sndBump       = 7 // walking into a wall
	sndSpell51    = 8
	sndMiss       = 9 // an attack hits nobody
	sndStep       = 10
	sndSpell47    = 11 // also the script call 0x3201 in areas of kind 10
	sndMissileEnd = 12 // a missile arrives
	sndTitle      = 13 // the Forgotten Realms banner
	sndCountdown  = 15
)

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
	spellAnimateDead                   = 36
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
	spellSpell57                       = 57
	spellCureSeriousWounds             = 58
	spellSpell59                       = 59
	spellSpell60                       = 60
	spellSpell61                       = 61
	spellSpell62                       = 62
	spellSpell63                       = 63
	spellSpell64                       = 64
	spellSpell65                       = 65
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
	spellSpell89                       = 89
	spellSpell90                       = 90
	spellCloudKill                     = 91
	spellConeOfCold                    = 92
	spellFeeblemind                    = 93
	spellHoldMonsters                  = 94
	spellSpell95                       = 95
	spellSpell96                       = 96
	spellSpell97                       = 97
	spellSpell98                       = 98
	spellSpell99                       = 99
	spellBestowCurseMU                 = 100
)

// MusicPlaces: every distinct musicPlace trigger, sorted.
func MusicPlaces() []string {
	var l []string
	for _, v := range musicPlace {
		if !slices.Contains(l, v) {
			l = append(l, v)
		}
	}
	slices.Sort(l)
	return l
}

// The temple's services (DS:00BD).
var templeServices = [10]string{"Cure Blindness", "Cure Disease", "Cure Light Wounds", "Cure Serious Wounds",
	"Cure Critical Wounds", "Heal", "Neutralize Poison", "Raise Dead", "Remove Curse", "Stone to Flesh"}

// templePrices are the services' prices in gold, as templeServices.
var templePrices = [10]uint16{1000, 1000, 100, 350, 600, 5000, 1000, 5500, 3500, 2000}

// templeNot are the cures' own refusals (templeCure, castCureAnyway), by
// service; the wound cures and Heal have none.
var templeNot = [10]string{0: "is not blind.", 1: "is not Diseased.", 6: "is not poisoned.", 7: "is not dead.", 8: "is not cursed.", 9: "is not stoned."}

// Title pictures in TITLE.DAX (Data.Art's title pieces).
const ()

// Item powers itemReadyEffect knows beyond classcalc.go's.
const (
	effectGrantsAffect   = 0 // the affect Spells[1] names
	effectAlignmentBurn  = 4 // Spells[1]: alignment (low nibble), damage (high)
	effectStrength       = 5
	effectConstitution   = 6
	effectScore          = 8 // Spells[1]: the score
	effectHammer         = 9 // unready: the spiritual hammer goes
	effectDexterity      = 10
	effectIntelligence   = 12
	effectStrAndInt      = 13
	alignmentMask        = 15
	alignmentDamageShift = 4
	wizardrySlots        = 3 // the first mage slot levels the ring doubles
	mageSlotLevels       = 5
)

// musicPlace: the ECL block (unique across ECL1..6) → its custom-music
// place trigger, named after what the block's own texts call it.
var musicPlace = map[byte]string{
	1: "tilverton", 2: "tilverton", 3: "tilverton-sewers", 4: "fire-knives-hideout",
	16: "pit-of-moander", 17: "yulash", 18: "pit-of-moander", 19: "yulash", 20: "yulash", 21: "pit-of-moander",
	32: "zhentil-keep", 33: "zhentil-keep-arena", 34: "zhentil-keep", 35: "zhentil-keep-arena", 36: "zhentil-keep", 37: "zhentil-keep",
	48: "haptooth-hills", 49: "haptooth-village", 50: "haptooth-hills", 51: "dracandros-tower", 52: "haptooth-hills", 53: "drow-caves",
	64: "myth-drannor", 65: "myth-drannor", 66: "temple-of-bane", 67: "temple-of-bane", 68: "myth-drannor", 69: "myth-drannor",
	80: "the-dalelands", 81: "the-dalelands", 82: "the-dalelands",
}
