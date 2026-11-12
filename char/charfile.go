package char

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"gobox/core"
	"rules1e"
)

// The rules' records in their file layouts (encoding/binary): the items of
// .itm and ITEM blocks, the affects of .spc and MON SPC blocks, the ITEMS
// table. The rules hold them as their state; this is where the files name them.
type (
	Item         = rules1e.Item
	AffectRecord = rules1e.AffectRecord
	ItemBase     = rules1e.ItemBase
)

// The record sizes of Item and AffectRecord.
const (
	ItemSize   = rules1e.ItemRecSize
	AffectSize = rules1e.AffectRecSize
)

// Files is one Pool of Radiance character as saved, each file's bytes (nil:
// the file is not there).
type Files struct {
	Record   []byte // .cha (roster) or .sav (game): the 0x11D record, then the pool trailer
	Items    []byte // .itm
	Affects  []byte // .spc
	Plan     []byte // .mem
	Dual     []byte // .dual
	Powers   []byte // .pow
	Portrait []byte // .png
	Icon     []byte // .icon.png
}

// Pack285 is f as the canonical character (rules1e.CharacterFile's JSON).
func Pack285(f Files) (core.CharacterFile, error) {
	c := rules1e.CharacterJSON{Sheet: SheetOf(f.Record), PoolIDs: ReadPoolIDs(f.Record), Portrait: f.Portrait, Icon: f.Icon}
	for _, it := range ReadRecords[Item](f.Items, ItemSize) {
		c.Items = append(c.Items, *it)
	}
	for _, a := range ReadRecords[AffectRecord](f.Affects, AffectSize) {
		c.Affects = append(c.Affects, *a)
	}
	if f.Plan != nil {
		c.Plan = new([21]rules1e.SpellID)
		fromBytes(c.Plan[:], f.Plan)
	}
	fromBytes(c.OldLevels[:], f.Dual)
	readPowers(&c, f.Powers)
	b, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("Pack285: %w", err)
	}
	return b, nil
}

// Unpack285 is the canonical character as its files; a game save's items
// file (game) carries the shared pack the party's first member holds.
func Unpack285(b core.CharacterFile, game bool) (Files, error) {
	var c rules1e.CharacterJSON
	if err := json.Unmarshal(b, &c); err != nil {
		return Files{}, fmt.Errorf("Unpack285: %w", err)
	}
	items := c.Items
	if game {
		items = append(items, c.Pool...)
	}
	f := Files{Record: append(SheetBytes(c.Sheet), PoolTrailer(c.PoolIDs)...), Portrait: c.Portrait, Icon: c.Icon, Powers: powerBytes(c)}
	f.Items = WriteRecords(pointers(items))
	f.Affects = WriteRecords(pointers(c.Affects))
	if c.Plan != nil {
		f.Plan = bytes.TrimRight(fileBytes(c.Plan[:]), "\x00")
	}
	f.Dual = bytes.TrimRight(fileBytes(c.OldLevels[:]), "\x00")
	return f, nil
}

func pointers[T any](v []T) []*T {
	var out []*T
	for i := range v {
		out = append(out, &v[i])
	}
	return out
}

// Monster285 is a MON block's records for the rules.
func Monster285(r Record285, affects []AffectRecord, items []Item) rules1e.Monster {
	return rules1e.Monster{Sheet: sheetOf(r), Affects: affects, Items: items}
}

// SheetOf reads a record in the 0x11D layout (short data leaves the rest zero).
func SheetOf(b []byte) rules1e.Sheet { return sheetOf(Decode[Record285](b)) }

// SheetBytes is the sheet as a 0x11D record (pointers zero).
func SheetBytes(s rules1e.Sheet) []byte { return Encode(recordOf(s)) }

// ReadRecords reads the n-byte records in b (a partial last one is dropped).
func ReadRecords[T any](b []byte, n int) []*T {
	var out []*T
	for ; len(b) >= n; b = b[n:] {
		v := new(T)
		binary.Read(bytes.NewReader(b[:n]), binary.LittleEndian, v)
		out = append(out, v)
	}
	return out
}

// WriteRecords is the inverse of ReadRecords.
func WriteRecords[T any](list []*T) []byte {
	var buf bytes.Buffer
	for _, v := range list {
		binary.Write(&buf, binary.LittleEndian, v)
	}
	return buf.Bytes()
}

// fileBytes is v as the bytes of a save file (.mem, .dual).
func fileBytes[T ~uint8](v []T) []byte {
	b := make([]byte, len(v))
	bytesOf(b, v)
	return b
}

// fromBytes fills v from b (a save file's bytes), as copy does.
func fromBytes[T ~uint8](v []T, b []byte) {
	for i := range min(len(v), len(b)) {
		v[i] = T(b[i])
	}
}

// powerBytes are c's Powers for NAME.pow: the ids, then a 0 and the
// daily uses spent (Surges, Laid) and the Toughness in MaxHP (Tough), then
// the reactions' policies.
func powerBytes(c rules1e.CharacterJSON) []byte {
	var b []byte
	for _, id := range c.Powers {
		b = append(b, byte(id))
	}
	b = append(b, 0, byte(c.Surges), byte(c.Laid), byte(c.Tough))
	for _, id := range slices.Sorted(maps.Keys(c.Policies)) {
		b = append(b, byte(id), byte(c.Policies[id]))
	}
	return b
}

// readPowers reads NAME.pow's bytes into c's Powers and daily uses.
func readPowers(c *rules1e.CharacterJSON, b []byte) {
	for i, id := range b {
		if id == 0 { // the daily uses (an older file has none)
			if i+2 < len(b) {
				c.Surges, c.Laid = int(b[i+1]), int(b[i+2])
			}
			if i+3 < len(b) {
				c.Tough = int(b[i+3])
			}
			for j := i + 5; j < len(b); j += 2 { // the reactions' policies (an older file has none)
				if c.Policies == nil {
					c.Policies = map[int]rules1e.Policy{}
				}
				c.Policies[int(b[j-1])] = rules1e.Policy(b[j])
			}
			return
		}
		c.Powers = append(c.Powers, int(id))
	}
}

// readStr is the length-prefixed string at the start of t and the rest.
func readStr(t []byte) (string, []byte, bool) {
	if len(t) < 2 || len(t) < 2+int(binary.LittleEndian.Uint16(t)) {
		return "", nil, false
	}
	n := 2 + int(binary.LittleEndian.Uint16(t))
	return string(t[2:n]), t[n:], true
}

// ReadPoolIDs parses the trailer of a character file image; zero when none.
func ReadPoolIDs(b []byte) (p rules1e.PoolIDs) {
	n := int(Layout285)
	if len(b) <= n || !bytes.HasPrefix(b[n:], []byte(rules1e.PoolMagic)) {
		return
	}
	f := p.Fields()
	for t := b[n+len(rules1e.PoolMagic):]; len(t) > 0; {
		k, rest, ok := readStr(t)
		if !ok {
			break
		}
		v, rest, ok := readStr(rest)
		if !ok {
			break
		}
		if dst := f[k]; dst != nil {
			*dst = core.PicID(v)
		} else if n, err := strconv.ParseUint(v, 10, 32); k == rules1e.RosterKey && err == nil {
			p.Roster = core.RosterID(n)
		}
		t = rest
	}
	return
}

func appendStr(b []byte, s string) []byte {
	return append(binary.LittleEndian.AppendUint16(b, uint16(len(s))), s...)
}

// PoolTrailer is p as the trailer after a record (nil: nothing to keep).
func PoolTrailer(p rules1e.PoolIDs) []byte {
	var b []byte
	f := p.Fields()
	for _, k := range rules1e.PoolKeys {
		if v := *f[k]; v != "" {
			b = appendStr(appendStr(b, k), string(v))
		}
	}
	if p.Roster != core.NoRoster {
		b = appendStr(appendStr(b, rules1e.RosterKey), strconv.FormatUint(uint64(p.Roster), 10))
	}
	if b == nil {
		return nil
	}
	return append([]byte(rules1e.PoolMagic), b...)
}

// sheetOf is the record's character fields (the name up to its 15 letters).
func sheetOf(r Record285) rules1e.Sheet {
	n := min(int(r.PName[0]), 15)
	return rules1e.Sheet{
		CharName:   string(r.PName[1 : 1+n]),
		Pad86:      r.Pad86[0],
		Pad87:      r.Pad86[1],
		Stats:      r.Stats,
		Str00:      r.Str00,
		Memorized:  r.Memorized,
		Study:      r.Study,
		Thac0Base:  r.Thac0Base,
		Race:       r.Race,
		Class:      r.Class,
		Age:        r.Age,
		MaxHP:      r.MaxHP,
		SpellBook:  r.SpellBook,
		FightLevel: r.FightLevel,
		Size:       r.Size,
		Saves:      r.Saves,
		BaseMove:   r.BaseMove,
		MaxLevel:   r.MaxLevel,
		Drained:    r.Drained,
		DrainedHP:  r.DrainedHP,
		Undead:     r.Undead,
		Thief:      r.Thief,
		Pad83:      r.Pad83,
		Control:    r.Control,
		Share:      r.Share,
		Coins:      r.Coins,
		Levels:     r.Levels,
		Sex:        r.Sex,
		Kind:       r.Kind,
		Alignment:  r.Alignment,
		BaseAttack: r.BaseAttack,
		BaseDice:   r.BaseDice,
		BaseSides:  r.BaseSides,
		BaseBonus:  r.BaseBonus,
		BaseAC:     r.BaseAC,
		StrBonus:   r.StrBonus,
		RandomID:   r.RandomID,
		Exp:        r.Exp,
		ClassMask:  r.ClassMask,
		HPRoll:     r.HPRoll,
		SpellSlots: slotsOf285(r.SpellSlots),
		ExpValue:   r.ExpValue,
		ExpPerHP:   r.ExpPerHP,
		PortHead:   r.PortHead,
		PortBody:   r.PortBody,
		IconHead:   r.IconHead,
		IconBody:   r.IconBody,
		Order:      r.Order,
		IconSize:   r.IconSize,
		IconColour: r.IconColour,
		ItemCount:  r.ItemCount,
		Hands:      r.Hands,
		SaveBonus:  r.SaveBonus,
		Weight:     r.Weight,
		Status:     r.Status,
		Okay:       r.Okay,
		Side:       r.Side,
		Auto:       r.Auto,
		THAC0:      r.THAC0,
		AC:         r.AC,
		Attacks:    r.Attacks,
		Dice:       r.Dice,
		Sides:      r.Sides,
		Bonus:      r.Bonus,
		HP:         r.HP,
		Move:       r.Move,
	}
}

// recordOf is the sheet in the record's layout (pointers zero).
func recordOf(s rules1e.Sheet) Record285 {
	r := Record285{
		Pad86:      [2]byte{s.Pad86, s.Pad87},
		Stats:      s.Stats,
		Str00:      s.Str00,
		Memorized:  s.Memorized,
		Study:      s.Study,
		Thac0Base:  s.Thac0Base,
		Race:       s.Race,
		Class:      s.Class,
		Age:        s.Age,
		MaxHP:      s.MaxHP,
		SpellBook:  s.SpellBook,
		FightLevel: s.FightLevel,
		Size:       s.Size,
		Saves:      s.Saves,
		BaseMove:   s.BaseMove,
		MaxLevel:   s.MaxLevel,
		Drained:    s.Drained,
		DrainedHP:  s.DrainedHP,
		Undead:     s.Undead,
		Thief:      s.Thief,
		Pad83:      s.Pad83,
		Control:    s.Control,
		Share:      s.Share,
		Coins:      s.Coins,
		Levels:     s.Levels,
		Sex:        s.Sex,
		Kind:       s.Kind,
		Alignment:  s.Alignment,
		BaseAttack: s.BaseAttack,
		BaseDice:   s.BaseDice,
		BaseSides:  s.BaseSides,
		BaseBonus:  s.BaseBonus,
		BaseAC:     s.BaseAC,
		StrBonus:   s.StrBonus,
		RandomID:   s.RandomID,
		Exp:        s.Exp,
		ClassMask:  s.ClassMask,
		HPRoll:     s.HPRoll,
		SpellSlots: slots285(s.SpellSlots),
		ExpValue:   s.ExpValue,
		ExpPerHP:   s.ExpPerHP,
		PortHead:   s.PortHead,
		PortBody:   s.PortBody,
		IconHead:   s.IconHead,
		IconBody:   s.IconBody,
		Order:      s.Order,
		IconSize:   s.IconSize,
		IconColour: s.IconColour,
		ItemCount:  s.ItemCount,
		Hands:      s.Hands,
		SaveBonus:  s.SaveBonus,
		Weight:     s.Weight,
		Status:     s.Status,
		Okay:       s.Okay,
		Side:       s.Side,
		Auto:       s.Auto,
		THAC0:      s.THAC0,
		AC:         s.AC,
		Attacks:    s.Attacks,
		Dice:       s.Dice,
		Sides:      s.Sides,
		Bonus:      s.Bonus,
		HP:         s.HP,
		Move:       s.Move,
	}
	r.PName[0] = byte(copy(r.PName[1:], s.CharName[:min(len(s.CharName), 15)]))
	return r
}

// slotsOf285 and slots285 move the 0x11D record's slots (cleric 1..3,
// magic-user 1..3) to and from rules1e's by spell class and level; the
// record has no room for the rest.
func slotsOf285(r [6]rules1e.Count) (s rules1e.SpellSlots) {
	copy(s[rules1e.SpellClassCleric][:3], r[0:3])
	copy(s[rules1e.SpellClassMage][:3], r[3:6])
	return s
}

func slots285(s rules1e.SpellSlots) (r [6]rules1e.Count) {
	copy(r[0:3], s[rules1e.SpellClassCleric][:3])
	copy(r[3:6], s[rules1e.SpellClassMage][:3])
	return r
}
