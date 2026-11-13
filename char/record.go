package char

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"

	"gobox/rules"
	"gobox/rules1e"
)

// The layouts (core.RecordLayout), here for the codecs' users.
const (
	Layout285 = rules.Layout285
	Layout422 = rules.Layout422
	Layout439 = rules.Layout439
	Layout510 = rules.Layout510
)

// Record is one of the layouts' structs.
type Record interface {
	Record285 | Record422 | Record439 | Record510
}

// Decode reads a record from b (short data leaves the rest zero; pointer
// fields are not kept).
func Decode[R Record](b []byte) R {
	var r R
	buf := make([]byte, binary.Size(r))
	copy(buf, b)
	binary.Read(bytes.NewReader(buf), binary.LittleEndian, &r) // buf has the exact size: cannot fail
	return r
}

// Encode is r in its layout (pointers zero).
func Encode[R Record](r R) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, &r) // fixed-size struct: cannot fail
	return buf.Bytes()
}

// Kept is b (one record) with the bytes of R's blank fields (the DOS
// pointers) zeroed: what Encode(Decode(b)) gives back.
func Kept[R Record](b []byte) []byte {
	out := bytes.Clone(b)
	t := reflect.TypeFor[R]()
	off := 0
	for i := range t.NumField() {
		f := t.Field(i)
		n := binary.Size(reflect.New(f.Type).Elem().Interface())
		if f.Name == "_" {
			clear(out[off : off+n])
		}
		off += n
	}
	return out
}

// StatPair is one ability score: the character's own (Base) and the value
// the rules use.
type StatPair struct {
	Base, Value byte
}

// Str00Pair is exceptional strength: the value in use and the character's
// own, the reverse of StatPair.
type Str00Pair struct {
	Value, Base byte
}

// Record285 is Pool of Radiance's 0x11D-byte record (.SAV files, MON*CHA
// blocks, the imported data's JSON).
type Record285 struct {
	PName      [16]byte               // 0x00 Pascal string, max 15
	Stats      [6]rules1e.Score       // 0x10 str int wis dex con cha
	Str00      byte                   // 0x16 exceptional strength
	Memorized  [21]byte               // 0x17 spell numbers, +0x80 = still to be memorised (rules1e.SpellOfByte)
	Study      byte                   // 0x2C hours of study the pending spells need (0/4/6)
	Thac0Base  byte                   // 0x2D THAC0 base, from the DS:3C7C table
	Race       rules1e.RaceID         // 0x2E
	Class      rules1e.ClassID        // 0x2F
	Age        uint16                 // 0x30
	MaxHP      byte                   // 0x32
	SpellBook  [56]byte               // 0x33 [n-1]: spell n known
	FightLevel byte                   // 0x6B fighter level (1 for non-fighters)
	Size       byte                   // 0x6C low bits: icon size; 0x80 flag
	Saves      [5]rules1e.Score       // 0x6D saving throw targets
	BaseMove   byte                   // 0x72
	MaxLevel   byte                   // 0x73 highest class level
	Kept74     byte                   // 0x74 kept for byte-equality (was the level-drain count; no game drains)
	Kept75     byte                   // 0x75 kept for byte-equality (was the drained hit points)
	Undead     byte                   // 0x76 turn undead row
	Thief      [8]rules1e.Percent     // 0x77 thief skills
	_          [4]byte                // 0x7F affect list pointer
	Pad83      byte                   // 0x83
	Control    byte                   // 0x84 0 party, 0xB3 charmed, > 0x7F hired NPC
	Share      byte                   // 0x85 hired NPC's treasure shares (& 7)
	Pad86      [2]byte                // 0x86
	Coins      [7]uint16              // 0x88 copper .. jewelry: only read when joining the party
	Levels     [8]rules1e.Level       // 0x96 cleric druid fighter paladin ranger mage thief monk
	Sex        byte                   // 0x9E
	Kind       byte                   // 0x9F monster kind (bonuses against it)
	Alignment  byte                   // 0xA0
	BaseAttack [2]rules1e.Count       // 0xA1 attacks per hand
	BaseDice   [2]rules1e.Count       // 0xA3 damage dice count per hand
	BaseSides  [2]rules1e.Count       // 0xA5 damage dice sides per hand
	BaseBonus  [2]rules1e.Modifier    // 0xA7 damage bonus per hand
	BaseAC     byte                   // 0xA9
	StrBonus   byte                   // 0xAA non-zero: strength modifies hit and damage
	RandomID   byte                   // 0xAB
	Exp        uint32                 // 0xAC
	ClassMask  byte                   // 0xB0 items the classes may use
	HPRoll     byte                   // 0xB1 rolled hit points
	SpellSlots [6]rules1e.Count       // 0xB2 cleric levels 1..3, magic-user levels 1..3
	ExpValue   uint16                 // 0xB8 a monster's worth
	ExpPerHP   byte                   // 0xBA
	PortHead   byte                   // 0xBB
	PortBody   byte                   // 0xBC
	IconHead   byte                   // 0xBD
	IconBody   byte                   // 0xBE
	Order      byte                   // 0xBF
	IconSize   byte                   // 0xC0
	IconColour [6]rules1e.Colours     // 0xC1 icon parts, low nibble and high nibble
	ItemCount  byte                   // 0xC7
	_          [4]byte                // 0xC8 item list pointer
	_          [52]byte               // 0xCC readied item pointers
	Hands      byte                   // 0x100
	SaveBonus  byte                   // 0x101
	Weight     uint16                 // 0x102
	_          [8]byte                // 0x104 party link and combat block pointers
	Status     byte                   // 0x10C index into DS:1047 names
	Okay       byte                   // 0x10D non-zero = active
	Side       byte                   // 0x10E 0 party, 1 monsters (1 also draws the name yellow)
	Auto       byte                   // 0x10F the computer fights for it
	THAC0      byte                   // 0x110
	AC         [2]rules1e.ArmourClass // 0x111 60 - armour class: front, rear
	Attacks    [2]rules1e.Count       // 0x113 attacks per hand this round
	Dice       [2]rules1e.Count       // 0x115 damage dice count per hand
	Sides      [2]rules1e.Count       // 0x117 damage dice sides per hand
	Bonus      [2]rules1e.Modifier    // 0x119 damage bonus per hand
	HP         byte                   // 0x11B
	Move       byte                   // 0x11C
}

// Record422 is Curse of the Azure Bonds' 422-byte record (CHRDAT/SAVGAM, MON*CHA). Offsets decimal.
type Record422 struct {
	PName           [16]byte    // 0 Pascal string, max 15
	Stats           [6]StatPair // 16 str int wis dex con cha
	Str00           Str00Pair   // 28 exceptional strength
	Memorized       [84]byte    // 30 spell ids, +128 = still to be memorised
	SpellsToLearn   byte        // 114
	THAC0Base       byte        // 115
	Race            byte        // 116 (7 = human)
	Class           byte        // 117
	Age             uint16      // 118
	MaxHP           byte        // 120
	SpellBook       [100]byte   // 121 [n-1]: spell n known
	FightLevel      byte        // 221 attack level (1 for non-fighters)
	Size            byte        // 222 low bits: icon size; 128 flag
	Saves           [5]byte     // 223 saving throw targets
	BaseMove        byte        // 228
	MaxLevel        byte        // 229 highest class level (hit dice)
	MulticlassLevel byte        // 230
	Kept74          byte        // 231 kept for byte-equality (was the level-drain count; no game drains)
	Kept75          byte        // 232 kept for byte-equality (was the drained hit points)
	Undead          byte        // 233 turn undead row
	Thief           [8]byte     // 234 thief skills
	_               [4]byte     // 242 affect list pointer
	Pad83           byte        // 246
	Control         byte        // 247 0 party, 179 charmed, above 127 hired NPC
	Share           byte        // 248 hired NPC's treasure shares (& 7)
	Pad86           [2]byte     // 249
	Coins           [7]uint16   // 251 copper .. jewelry: only read when joining the party
	Levels          [8]byte     // 265 cleric druid fighter paladin ranger mage thief monk
	OldLevels       [8]byte     // 273 a dual-classed human's levels in the class left behind
	Sex             byte        // 281
	Kind            byte        // 282 monster kind (bonuses against it)
	Alignment       byte        // 283
	BaseAttack      [2]byte     // 284 attacks per hand
	BaseDice        [2]byte     // 286 damage dice count per hand
	BaseSides       [2]byte     // 288 damage dice sides per hand
	BaseBonus       [2]byte     // 290 damage bonus per hand
	BaseAC          byte        // 292
	StrBonus        byte        // 293 non-zero: strength modifies hit and damage
	RandomID        byte        // 294
	Exp             uint32      // 295
	ClassMask       byte        // 299 items the classes may use
	HPRoll          byte        // 300 rolled hit points
	SpellSlots      [15]byte    // 301 spells per level: cleric 1..5, druid 1..5, magic-user 1..5
	ExpValue        uint16      // 316 a monster's worth
	ExpPerHP        byte        // 318
	PortHead        byte        // 319
	PortBody        byte        // 320
	IconHead        byte        // 321
	IconBody        byte        // 322
	Order           byte        // 323
	IconSize        byte        // 324
	IconColour      [6]byte     // 325 icon parts, low nibble and high nibble
	Pad331          byte        // 331
	ItemCount       byte        // 332
	_               [4]byte     // 333 item list pointer
	_               [52]byte    // 337 readied item pointers
	Hands           byte        // 389
	SaveBonus       byte        // 390
	Weight          uint16      // 391
	_               [8]byte     // 393 party link and combat block pointers
	PaladinCures    byte        // 401 cure disease uses left this week
	Pad402          [3]byte     // 402
	Status          byte        // 405 index into the status names
	Okay            byte        // 406 non-zero = active
	Side            byte        // 407 0 party, 1 monsters (1 also draws the name yellow)
	Auto            byte        // 408 the computer fights for it
	THAC0           byte        // 409
	AC              [2]byte     // 410 60 - armour class: front, rear
	Attacks         [2]byte     // 412 attacks per hand this round
	Dice            [2]byte     // 414 damage dice count per hand
	Sides           [2]byte     // 416 damage dice sides per hand
	Bonus           [2]byte     // 418 damage bonus per hand
	HP              byte        // 420
	Move            byte        // 421
}

// Record439 is Secret of the Silver Blades' 439-byte record (CHRDATA/SAVGAM, MON*CHA).
type Record439 struct {
	PName           [16]byte    // 0 Pascal string, max 15
	Stats           [6]StatPair // 16 str int wis dex con cha
	Str00           Str00Pair   // 28 exceptional strength
	Memorized       [75]byte    // 30 spell ids, +128 = still to be memorised
	SpellsToLearn   byte        // 105
	THAC0Base       byte        // 106
	Race            byte        // 107 (raceHuman ...)
	Class           byte        // 108
	PaladinCures    byte        // 109 cure disease uses left this week
	Age             uint16      // 110
	MaxHP           byte        // 112
	SpellBook       [117]byte   // 113 [n-1]: spell n known
	FightLevel      byte        // 230 attack level (1 for non-fighters)
	Size            byte        // 231 low bits: icon size; 128 flag
	Saves           [5]byte     // 232 saving throw targets
	BaseMove        byte        // 237
	MaxLevel        byte        // 238 highest class level (hit dice)
	MulticlassLevel byte        // 239
	Kept74          byte        // 240 kept for byte-equality (was the level-drain count; no game drains)
	Kept75          byte        // 241 kept for byte-equality (was the drained hit points)
	Undead          byte        // 242 turn undead row
	Thief           [8]byte     // 243 thief skills
	_               [4]byte     // 251 affect list pointer
	Control         byte        // 255 0 party, 179 charmed, above 127 hired NPC
	Share           byte        // 256 hired NPC's treasure shares (& 7)
	OldClass        byte        // 257 the class a dual-classed human left (clsCleric ...)
	Pad258          byte        // 258
	Coins           [7]uint16   // 259 copper .. jewelry: only read when joining the party
	Levels          [7]byte     // 273 cleric druid fighter paladin ranger mage thief
	OldLevels       [7]byte     // 280 a dual-classed human's levels in the class left behind
	Sex             byte        // 287
	Alignment       byte        // 288
	BaseAttack      [2]byte     // 289 attacks per hand
	BaseDice        [2]byte     // 291 damage dice count per hand
	BaseSides       [2]byte     // 293 damage dice sides per hand
	BaseBonus       [2]byte     // 295 damage bonus per hand
	BaseAC          byte        // 297
	StrBonus        byte        // 298 non-zero: strength modifies hit and damage
	RandomID        byte        // 299
	Exp             uint32      // 300
	ClassMask       byte        // 304 items the classes may use
	HPRoll          byte        // 305 rolled hit points
	SpellSlots      [28]byte    // 306 spells per level, seven levels each: cleric, druid, paladin/ranger, magic-user
	ExpValue        uint16      // 334 a monster's worth
	ExpPerHP        byte        // 336
	PortHead        byte        // 337
	PortBody        byte        // 338
	IconHead        byte        // 339
	IconBody        byte        // 340
	Order           byte        // 341
	IconSize        byte        // 342
	IconColour      [6]byte     // 343 icon parts, low nibble and high nibble
	CreatureFlags   byte        // 349 creature kind bits (creatureEnchanted ...)
	MonsterFlags    byte        // 350 more kind bits
	Pad351          byte        // 351
	ItemCount       byte        // 352
	_               [4]byte     // 353 item list pointer
	_               [52]byte    // 357 readied item pointers
	Hands           byte        // 409
	SaveBonus       byte        // 410
	Weight          uint16      // 411
	_               [8]byte     // 413 party link and combat block pointers
	MagicResistance byte        // 421 percent, before the caster level
	Status          byte        // 422 index into the status names
	Okay            byte        // 423 non-zero = active
	Side            byte        // 424 0 party, 1 monsters (1 also draws the name yellow)
	Auto            byte        // 425 the computer fights for it
	THAC0           byte        // 426
	AC              [2]byte     // 427 60 - armour class: front, rear
	Attacks         [2]byte     // 429 attacks per hand this round
	Dice            [2]byte     // 431 damage dice count per hand
	Sides           [2]byte     // 433 damage dice sides per hand
	Bonus           [2]byte     // 435 damage bonus per hand
	HP              byte        // 437
	Move            byte        // 438
}

// Record510 is Pools of Darkness' 510-byte record (.PC, .SAV, monsters).
type Record510 struct {
	PName           [16]byte    // 0 Pascal string, max 15
	Stats           [6]StatPair // 16 str int wis dex con cha
	Str00           Str00Pair   // 28 exceptional strength
	Memorized       [141]byte   // 30 spell ids, +128 = still to be memorised
	SpellsToLearn   byte        // 171
	THAC0Base       byte        // 172
	Race            byte        // 173 (raceHuman ...)
	Class           byte        // 174
	PaladinCures    byte        // 175 cure disease uses left this week
	Age             uint16      // 176
	MaxHP           byte        // 178
	SpellBook       [126]byte   // 179 [n-1]: spell n known
	Size            byte        // 305 low bits: icon size; 128 flag
	Saves           [5]byte     // 306 saving throw targets
	BaseMove        byte        // 311
	MaxLevel        byte        // 312 highest class level (hit dice)
	MulticlassLevel byte        // 313 a dual-classed human's level when changing class
	Undead          byte        // 314 turn undead row
	Thief           [8]byte     // 315 thief skills
	_               [4]byte     // 323 affect list pointer
	Control         byte        // 327 0 party, 179 charmed, above 127 hired NPC
	Share           byte        // 328 hired NPC's treasure shares (& 7)
	OldClass        byte        // 329 the class a dual-classed human left (clsCleric ...)
	Pad330          byte        // 330
	Coins           [3]uint16   // 331 platinum, gems, jewelry: only read when joining the party
	Levels          [7]byte     // 337 cleric druid fighter paladin ranger mage thief
	OldLevels       [7]byte     // 344 a dual-classed human's levels in the class left behind
	HighestLevels   [7]byte     // 351 levels before energy drain (restoration gives them back)
	Sex             byte        // 358
	Alignment       byte        // 359
	BaseAttack      [2]byte     // 360 attacks per hand
	BaseDice        [2]byte     // 362 damage dice count per hand
	BaseSides       [2]byte     // 364 damage dice sides per hand
	BaseBonus       [2]byte     // 366 damage bonus per hand
	BaseAC          byte        // 368
	RandomID        byte        // 369
	Exp             uint32      // 370
	HighestExp      uint32      // 374 experience before energy drain
	HighestHP       byte        // 378 maximum hit points before energy drain
	ClassMask       byte        // 379 items the classes may use
	HPRoll          byte        // 380 rolled hit points
	SpellSlots      [27]byte    // 381 spells per level, nine levels each: cleric, druid, magic-user
	ExpValue        uint16      // 408 a monster's worth
	IconHead        byte        // 410
	IconBody        byte        // 411
	Order           byte        // 412
	IconSize        byte        // 413
	IconColour      [6]byte     // 414 icon parts, low nibble and high nibble
	CreatureFlags   byte        // 420 creature kind bits (creatureEnchanted ...)
	MonsterFlags    byte        // 421 more kind bits
	ItemCount       byte        // 422
	_               [4]byte     // 423 item list pointer
	_               [52]byte    // 427 readied item pointers
	Hands           byte        // 479
	SaveBonus       byte        // 480
	Weight          uint16      // 481
	_               [8]byte     // 483 party link and combat block pointers
	MagicResistance byte        // 491 percent, before the caster level
	CanTrain        byte        // 492 non-zero: enough experience to train
	Status          byte        // 493 index into the status names
	Okay            byte        // 494 non-zero = active
	Side            byte        // 495 0 party, 1 monsters (1 also draws the name yellow)
	Auto            byte        // 496 the computer fights for it
	THAC0           byte        // 497
	AC              [2]byte     // 498 60 - armour class: front, rear
	Attacks         [2]byte     // 500 attacks per hand this round
	Dice            [2]byte     // 502 damage dice count per hand
	Sides           [2]byte     // 504 damage dice sides per hand
	Bonus           [2]byte     // 506 damage bonus per hand
	HP              byte        // 508
	Move            byte        // 509
}

func init() {
	for _, c := range []struct {
		l rules.RecordLayout
		n int
	}{{Layout285, binary.Size(Record285{})}, {Layout422, binary.Size(Record422{})}, {Layout439, binary.Size(Record439{})}, {Layout510, binary.Size(Record510{})}} {
		if int(c.l) != c.n {
			panic(fmt.Sprintf("char: layout %d record is %d bytes", c.l, c.n))
		}
	}
}
