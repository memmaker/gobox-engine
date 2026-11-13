package rules1e

import "gobox/rules"

// Character is a party member or a monster: its record (Sheet), its item
// and affect lists, its combat block and readied items.
type Character struct {
	Sheet
	Portrait        rules.PNG            // extra: the encoded PNG dropped on the sheet, shown everywhere in place of the paper-doll head (saved as NAME.png beside the record; nil: none)
	PoolIDs                              // extra: gobox's opaque pool picks (poolid.go; saved in the .cha/.sav trailer)
	Icon            rules.PNG            // extra: the encoded PNG dropped on the icon editor, the combat icon in battle (iconpng.go; saved as NAME.icon.png; nil: the built icon)
	Plan            [MemorySlots]SpellID // extra: the spells memorized on every rest (camp.go memorizeMenu; saved as NAME.mem)
	OldLevels       [8]Level             // extra: a dual-classed human's levels in the class left (dualclass.go; saved as NAME.dual)
	Powers          []int                // extra: the rule set's powers chosen (powers.go; saved as NAME.pow)
	Surges          int                  // extra: the rule set's healing surges spent since the last long rest (saved in NAME.pow)
	Tough           int                  // extra: the hit points of the rule set's Toughness in MaxHP now (saved in NAME.pow)
	Laid            int                  // extra: the rule set's lays on hands since the last long rest (saved in NAME.pow)
	Policies        map[int]Policy       // extra: the standing answer to each of the rule set's reactions, by power id (saved in NAME.pow; missing: Always)
	MulticlassLevel Level                // extra: COTAB/SotSB/POD's multi-class level (+230/+239/+313; carried, no POR rule reads it)
	Kept            *KeptRecord          // extra: the game record the character came from (nil: POR's 0x11D), its fields rules1e does not model
	CanTrain        bool                 // pod +492: enough experience to train in a class (GainExperience sets it, the coordinator's 04:05 superset)
	Items           []*Item              // list at +0xC8 (0x3F-byte nodes)
	Affects         []*Affect            // list at +0x7F (9-byte nodes)
	Combat          *CombatInfo          // block at +0x108 while fighting
	Slots           [14]*Item            // readied items, +0xCC (set by ApplyItems)
}

// Sheet is a character's record: the fields of the DOS record (the codec,
// por-go's charRecord, keeps its layout), without its pointers.
// Values keep their byte width: the rules rely on the wrap-around and the
// signed readings.
type Sheet struct {
	CharName    string               // 0x00 Pascal string, max 15
	Stats       [6]Score             // 0x10 str int wis dex con cha
	Str00       byte                 // 0x16 exceptional strength
	Memorized   [MemorySlots]SpellID // 0x17 spell ids (SpellPending: still to be memorised); POR's record has the first PORMemory
	Study       byte                 // 0x2C hours of study the pending spells need (0/4/6)
	Thac0Base   byte                 // 0x2D THAC0 base, from the DS:3C7C table
	Race        rules.Race           // 0x2E, as the race the record's byte names (RaceOfRecord)
	MonsterRace byte                 // 0x2E of a monster's record: the number it has where it names no player race (0 or another)
	Class       ClassID              // 0x2F
	Age         uint16               // 0x30
	MaxHP       byte                 // 0x32
	SpellBook   [LastSpellID]Flag    // 0x33 [n-1]: spell n known; POR's record has the first 56
	FightLevel  byte                 // 0x6B fighter level (1 for non-fighters)
	Size        byte                 // 0x6C low bits: icon size; 0x80 flag
	Saves       [5]Score             // 0x6D saving throw targets
	BaseMove    byte                 // 0x72
	MaxLevel    byte                 // 0x73 highest class level
	Kept74      byte                 // 0x74 kept for byte-equality (was the level-drain count; no game drains)
	Kept75      byte                 // 0x75 kept for byte-equality (was the drained hit points)
	Undead      byte                 // 0x76 turn undead row
	Thief       [8]Percent           // 0x77 thief skills
	Pad83       byte                 // 0x83
	Control     byte                 // 0x84 0 party, 0xB3 charmed, > 0x7F hired NPC
	Share       byte                 // 0x85 hired NPC's treasure shares (& 7)
	Pad86       byte                 // 0x86
	Pad87       byte                 // 0x87
	ScriptFlag  byte                 // COTAB record +402: the flags the scripts read (bit 0: imported from Hillsfar)
	Coins       [7]uint16            // 0x88 copper .. jewelry: only read when joining the party
	Levels      [8]Level             // 0x96 cleric druid fighter paladin ranger mage thief monk
	Sex         byte                 // 0x9E
	Kind        byte                 // 0x9F monster kind (bonuses against it)
	Alignment   byte                 // 0xA0
	BaseAttack  [2]Count             // 0xA1 attacks per hand
	BaseDice    [2]Count             // 0xA3 damage dice count per hand
	BaseSides   [2]Count             // 0xA5 damage dice sides per hand
	BaseBonus   [2]Modifier          // 0xA7 damage bonus per hand
	BaseAC      byte                 // 0xA9
	StrBonus    byte                 // 0xAA non-zero: strength modifies hit and damage
	RandomID    byte                 // 0xAB
	Exp         uint32               // 0xAC
	ClassMask   byte                 // 0xB0 items the classes may use
	HPRoll      byte                 // 0xB1 rolled hit points
	SpellSlots  SpellSlots           // 0xB2 in POR's record (cleric 1..3, magic-user 1..3); here by spell class and level 1..9, the codecs convert
	ExpValue    uint16               // 0xB8 a monster's worth
	ExpPerHP    byte                 // 0xBA
	PortHead    byte                 // 0xBB
	PortBody    byte                 // 0xBC
	IconHead    byte                 // 0xBD
	IconBody    byte                 // 0xBE
	Order       byte                 // 0xBF
	IconSize    byte                 // 0xC0
	IconColour  [6]Colours           // 0xC1 icon parts, low nibble and high nibble
	ItemCount   byte                 // 0xC7
	Hands       byte                 // 0x100
	SaveBonus   byte                 // 0x101
	Weight      uint16               // 0x102
	Status      byte                 // 0x10C index into DS:1047 names
	Okay        byte                 // 0x10D non-zero = active
	Side        byte                 // 0x10E 0 party, 1 monsters (1 also draws the name yellow)
	Auto        byte                 // 0x10F the computer fights for it
	THAC0       byte                 // 0x110
	AC          [2]ArmourClass       // 0x111 60 - armour class: front, rear
	Attacks     [2]Count             // 0x113 attacks per hand this round
	Dice        [2]Count             // 0x115 damage dice count per hand
	Sides       [2]Count             // 0x117 damage dice sides per hand
	Bonus       [2]Modifier          // 0x119 damage bonus per hand
	HP          byte                 // 0x11B
	Move        byte                 // 0x11C

	CreatureFlags   CreatureFlags `json:",omitempty"` // extra: CotAB +331 / SotSB +349 / PoD +420, what kind of foe it is (POR: none)
	MonsterFlags    MonsterFlags  `json:",omitempty"` // extra: SotSB +350 / PoD +421, more kind bits (POR: none)
	MagicResistance byte          `json:",omitempty"` // extra: SotSB +421 / PoD +491 percent, before the caster level (MagicResistRule); POR, CotAB none
	Base            [6]Score      `json:",omitempty"` // extra: CotAB/SotSB/PoD the own scores, before items (the record\'s score base); zero: Stats (POR)
	Str00Base       byte          `json:",omitempty"` // extra: the own exceptional strength
	PaladinCures    byte          `json:",omitempty"` // extra: CotAB/SotSB cures left this week (paladinCuresCotab); POR none
}

// The record's value types: each keeps the DOS byte's width and wrap-around.
type (
	Score       uint8  // an ability score or a saving throw target
	SpellID     uint16 // a unified spell id (spellids_games.go), SpellPending or'd in: still to be memorised
	Flag        uint8  // non-zero: set
	Percent     uint8  // a thief skill's chance
	Level       uint8  // a class level
	Count       uint8  // a count: attacks, dice, sides, spell slots, steps, rows
	Modifier    uint8  // a signed bonus (read as int8)
	Colours     uint8  // two icon colours, low and high nibble
	ArmourClass uint8  // 60 - armour class
	Slot        uint8  // a combatant index (Combatant), 0: none
	Ground      uint8  // a combat square's ground type (GroundTable)
	ClassID     uint8  // a class (ClsCleric..)
)

// Stats indexes.
const (
	StatStr = iota
	StatInt
	StatWis
	StatDex
	StatCon
	StatCha
)

// FighterKind: fighter, paladin or ranger levels; CotAB gives all three
// exceptional strength and the fighter's constitution hit points.
func FighterKind(c *Character) bool {
	return int8(c.Levels[ClsFighter]) > 0 || int8(c.Levels[ClsPaladin]) > 0 || int8(c.Levels[ClsRanger]) > 0
}

// Class values (the order of the DS:0D38 names); the first eight are also
// the Levels indexes.
const (
	ClsCleric = iota
	ClsDruid
	ClsFighter
	ClsPaladin
	ClsRanger
	ClsMage
	ClsThief
	ClsMonk
	clsClericFighter
	clsClericFighterMage
	ClsClericRanger
	clsClericMage
	clsClericThief
	clsFighterMage
	ClsFighterThief
	clsFighterMageThief
	clsMageThief
)

// Status values (the order of the DS:1047 names).
const (
	StatusOkay = iota
	StatusAnimated
	statusTempGone
	StatusRunning
	StatusUnconscious
	StatusDying
	StatusDead
	StatusStoned
	StatusGone
)

// Side values.
const (
	SideParty    = 0
	SideMonsters = 1
)

// SpellClass is a spell's class: a spell record's field 0 as Tables.Spells
// holds it. Each game's importer maps its own class byte to it (rules.json
// spellClasses); the first three index SpellSlots.
type SpellClass uint8

const (
	SpellClassCleric SpellClass = 0
	SpellClassMage   SpellClass = 1
	SpellClassDruid  SpellClass = 2
	SpellClassPower  SpellClass = 3 // an item's or monster's power: no slots
	SpellClasses                = 3 // the classes with slots
)

// MaxSpellLevel is the highest spell level of any game (POD's magic-user 9).
const MaxSpellLevel = 9

// SpellSlots are the spells a character can memorise, by spell class and level 1..9.
type SpellSlots [SpellClasses][MaxSpellLevel]Count

// SpellSlot is the number of spells of a level (1..9) a spell class can
// memorise; 0 for anything else.
func (c *Character) SpellSlot(class SpellClass, level int) Count {
	if class >= SpellClasses || level < 1 || level > MaxSpellLevel {
		return 0
	}
	return c.SpellSlots[class][level-1]
}

func (c *Character) Name() string {
	return c.CharName
}

// SetName stores a pascal string (max 15).
func (c *Character) SetName(s string) {
	if len(s) > 15 {
		s = s[:15]
	}
	c.CharName = s
}

// stats is a Character as gobox/rules reads it. POR keeps no multiclass level.
type stats struct{ c *Character }

func (s stats) Classes() int         { return len(s.c.Levels) }
func (s stats) Level(k int) int      { return int(s.c.Levels[k]) }
func (s stats) OldLevel(k int) int   { return int(s.c.OldLevels[k]) }
func (s stats) MulticlassLevel() int { return 0 }
func (s stats) HPRoll() int          { return int(s.c.HPRoll) }
func (s stats) MaxHP() int           { return int(s.c.MaxHP) }
func (s stats) HP() int              { return int(s.c.HP) }
func (s stats) THAC0() int           { return int(s.c.THAC0) }
func (s stats) AC(i int) int         { return int(s.c.AC[i]) }
func (s stats) Save(t int) int       { return int(s.c.Saves[t]) }
func (s stats) SaveBonus() int       { return int(int8(s.c.SaveBonus)) }
func (s stats) Status() int          { return int(s.c.Status) }
func (s stats) SetHP(v int)          { s.c.HP = byte(v) }
func (s stats) SetMaxHP(v int)       { s.c.MaxHP = byte(v) }
func (s stats) SetStatus(v int)      { s.c.Status = byte(v) }
func (s stats) SetOkay(v int)        { s.c.Okay = byte(v) }

// SpellPending marks a memory slot (or plan entry) whose spell is still to
// be memorised: the DOS byte's +0x80, moved above every unified id. The
// codecs (gobox/char) split a file byte into the id and this flag.
const SpellPending SpellID = 0x8000

// Memory slots: the Sheet holds the widest record's (POD's 141 bytes);
// the rules use POR's 21 (0x15), the rest stay as read.
const (
	MemorySlots = 141
	PORMemory   = 21
)

// SpellOfByte is a record's memory byte as a SpellID: the low 7 bits the
// spell number, +0x80 SpellPending. SpellID.Byte is its inverse.
func SpellOfByte(b byte) SpellID {
	s := SpellID(b & dosSpellBits)
	if b&dosPendingBit != 0 {
		s |= SpellPending
	}
	return s
}

// Byte is s as a record's memory byte (SpellOfByte's inverse); s.Spell()
// must be a game's spell number (< 0x80).
func (s SpellID) Byte() byte {
	b := byte(s.Spell())
	if s.Pending() {
		b |= dosPendingBit
	}
	return b
}

// The DOS memory and scroll byte: a spell number and the +0x80 flag.
const (
	dosSpellBits  = 0x7F
	dosPendingBit = 0x80
)

// Pending reports whether s is still to be memorised.
func (s SpellID) Pending() bool { return s&SpellPending != 0 }

// Spell is s without the pending flag.
func (s SpellID) Spell() SpellID { return s &^ SpellPending }

// Held reports whether s is a memorised spell (not empty, not pending).
func (s SpellID) Held() bool { return s != 0 && !s.Pending() }
