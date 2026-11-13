package rules1e

// Copied from cotab-go/engine/char.go and sotsb-go/engine/char.go (oracles).

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

type statPair struct{ Base, Value byte }
type str00Pair struct{ Value, Base byte }

type cotabRecord struct {
	PName           [16]byte    // 0 Pascal string, max 15
	Stats           [6]statPair // 16 str int wis dex con cha
	Str00           str00Pair   // 28 exceptional strength
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
type sotsbRecord struct {
	PName           [16]byte    // 0 Pascal string, max 15
	Stats           [6]statPair // 16 str int wis dex con cha
	Str00           str00Pair   // 28 exceptional strength
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

func TestLayoutsMatchStructs(t *testing.T) {
	for _, tc := range []struct {
		file string
		rec  any
	}{{"testdata/cotab.json", &cotabRecord{}}, {"testdata/sotsb.json", &sotsbRecord{}}} {
		l, err := LoadLayout(tc.file)
		if err != nil {
			t.Fatal(err)
		}
		if n := binary.Size(tc.rec); n != l.Size {
			t.Fatalf("%s: struct %d layout %d", tc.file, n, l.Size)
		}
		raw := make([]byte, l.Size)
		rand.New(rand.NewSource(1)).Read(raw)
		if err := binary.Read(bytes.NewReader(raw), binary.LittleEndian, tc.rec); err != nil {
			t.Fatal(err)
		}
		r := &Rec{B: raw, L: l}
		v := reflect.ValueOf(tc.rec).Elem()
		seen := 0
		get := func(name string) (reflect.Value, bool) {
			for i := 0; i < v.NumField(); i++ {
				fn := v.Type().Field(i).Name
				if strings.EqualFold(fn, name) || name == "name" && fn == "PName" || name == "class" && fn == "Class" {
					return v.Field(i), true
				}
			}
			return reflect.Value{}, false
		}
		for name, f := range l.Fields {
			var fv reflect.Value
			var ok bool
			sub := ""
			switch name {
			case "score", "scoreUse":
				fv, ok = get("Stats")
				sub = map[string]string{"score": "Base", "scoreUse": "Value"}[name]
			case "str00", "str00Use":
				fv, ok = get("Str00")
				sub = map[string]string{"str00": "Base", "str00Use": "Value"}[name]
			case "drained", "drainedHP": // the original's drain counters, kept as bytes (no game drains)
				fv, ok = get(map[string]string{"drained": "Kept74", "drainedHP": "Kept75"}[name])
			case "itemList", "slots":
				continue // pointers: not in the struct
			default:
				fv, ok = get(name)
			}
			if !ok {
				t.Errorf("%s: no struct field", name)
				continue
			}
			seen++
			n := max(f.N, 1)
			for i := 0; i < n; i++ {
				e := fv
				if fv.Kind() == reflect.Array {
					e = fv.Index(i)
				}
				if sub != "" {
					e = e.FieldByName(sub)
					if fv.Kind() != reflect.Array {
						e = fv.FieldByName(sub)
					}
				}
				if f.Size == 1 && f.N > 1 && fv.Kind() == reflect.Array && fv.Len() == n && fv.Index(0).Kind() == reflect.Uint8 && sub == "" {
					e = fv.Index(i)
				}
				want := int(e.Uint())
				if f.Signed {
					want = int(int8(want))
				}
				if got := r.Get(name, i); got != want {
					t.Errorf("%s: %s[%d] = %d want %d", tc.file, name, i, got, want)
				}
			}
		}
		if seen < 50 {
			t.Errorf("only %d fields checked", seen)
		}
	}
}

func TestSetGet(t *testing.T) {
	l, _ := LoadLayout("testdata/cotab.json")
	r := New(l)
	r.Set("exp", 0, 0x01020304)
	r.Set("hp", 0, 300)
	r.Set("saveBonus", 0, -3)
	if r.Get("exp", 0) != 0x01020304 || r.HP() != 44 || r.SaveBonus() != -3 || r.Classes() != 8 {
		t.Fatal("rec")
	}
	var _ CharStats = r
}
