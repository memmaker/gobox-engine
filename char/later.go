package char

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"reflect"

	"gobox/core"
	"rules1e"
)

// Later is a layout after Pool of Radiance's: COTAB (422), SotSB (439), POD (510).
type Later interface {
	Record422 | Record439 | Record510
}

// LaterFiles is a COTAB, SotSB or POD character as saved: the record, its
// items (ItemSize bytes each: 63, or 67 with a pointer) and its affects.
type LaterFiles struct {
	Record   []byte
	Items    []byte
	ItemSize int
	Affects  []byte
}

// PackLater is f as the canonical character: the fields rules1e models in
// the Sheet and the extras, the record itself kept for the rest. Spell ids
// (Memorized, SpellBook) stay in the kept record: rules1e's are POR's
// [21]/[56] by POR's numbers; they move once the unified spell list (S11)
// gives each game's number map.
func PackLater[R Later](f LaterFiles) (core.CharacterFile, error) {
	r := Decode[R](f.Record)
	c := rules1e.CharacterJSON{Kept: &rules1e.KeptRecord{Layout: core.RecordLayout(len(Encode(r))), Image: core.RecordImage(bytes.Clone(f.Record))}}
	sheetOfLater(&c, r)
	c.Items = readItems(f.Items, f.ItemSize)
	c.Affects = ReadAffects(f.Affects)
	b, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("PackLater: %w", err)
	}
	return b, nil
}

// UnpackLater is the canonical character as its files: the kept record with
// the modelled fields written over it (unchanged ones give the file back
// byte for byte, the DOS pointers included).
func UnpackLater[R Later](b core.CharacterFile, itemSize int) (LaterFiles, error) {
	var c rules1e.CharacterJSON
	if err := json.Unmarshal(b, &c); err != nil {
		return LaterFiles{}, fmt.Errorf("UnpackLater: %w", err)
	}
	var orig []byte
	if c.Kept != nil {
		orig = c.Kept.Image
	}
	r := Decode[R](orig)
	if c.Kept == nil || int(c.Kept.Layout) != len(Encode(r)) {
		return LaterFiles{}, fmt.Errorf("UnpackLater: kept record is not a %T", r)
	}
	laterOfSheet(&r, c)
	f := LaterFiles{Record: withBlanks[R](Encode(r), orig), ItemSize: itemSize}
	pad := make([]byte, itemSize-ItemSize)
	for _, it := range c.Items {
		f.Items = append(append(f.Items, WriteItems([]Item{it})...), pad...)
	}
	f.Affects = WriteAffects(c.Affects)
	return f, nil
}

// withBlanks is enc (an Encode) with R's blank fields' bytes (the DOS
// pointers) taken from orig, the inverse of Kept.
func withBlanks[R Record](enc, orig []byte) []byte {
	t := reflect.TypeFor[R]()
	off := 0
	for i := range t.NumField() {
		f := t.Field(i)
		n := int(f.Type.Size())
		if f.Name == "_" && off+n <= len(orig) {
			copy(enc[off:off+n], orig[off:off+n])
		}
		off += n
	}
	return enc
}

// The fields the codecs move by hand; every other field a later record and
// rules1e's Sheet both name is the same value in the same width and moves
// as it is (copySame).
var laterByHand = map[string]bool{
	"PName": true, "Stats": true, "Str00": true, "Memorized": true, "SpellBook": true,
	"SpellSlots": true, "Coins": true, "Pad86": true, "OldLevels": true,
	"MulticlassLevel": true, "CanTrain": true,
}

// laterAlias: a later record's name for a Sheet field, either way round.
var laterAlias = map[string]string{"THAC0Base": "Thac0Base", "Thac0Base": "THAC0Base"}

// copySame copies src's fields into dst's same-named ones: scalars, and
// byte arrays up to the shorter length (7 class levels into rules1e's 8).
func copySame(dst, src reflect.Value) {
	for i := range src.NumField() {
		f := src.Type().Field(i)
		if f.Name == "_" || laterByHand[f.Name] {
			continue
		}
		d := dst.FieldByName(cmp.Or(laterAlias[f.Name], f.Name))
		if !d.IsValid() {
			continue // not modelled: kept in the record
		}
		s := src.Field(i)
		if s.Kind() == reflect.Array {
			for j := range min(s.Len(), d.Len()) {
				d.Index(j).SetUint(s.Index(j).Uint())
			}
			continue
		}
		d.SetUint(s.Uint())
	}
}

// slotRuns are where a layout keeps each spell class's slots: offset into
// SpellSlots and levels (SotSB's paladin/ranger run, 14..20, has no rules1e
// class and stays in the record).
func slotRuns(r any) (runs [rules1e.SpellClasses]struct{ at, n int }) {
	set := func(cleric, druid, mage, n int) {
		runs[rules1e.SpellClassCleric] = struct{ at, n int }{cleric, n}
		runs[rules1e.SpellClassDruid] = struct{ at, n int }{druid, n}
		runs[rules1e.SpellClassMage] = struct{ at, n int }{mage, n}
	}
	switch r.(type) {
	case *Record422:
		set(0, 5, 10, 5)
	case *Record439:
		set(0, 7, 21, 7)
	case *Record510:
		set(0, 9, 18, 9)
	}
	return
}

// common are the hand-moved fields every later layout has.
type common struct {
	name  *[16]byte
	stats *[6]StatPair
	str00 *Str00Pair
	slots []byte
	coins []uint16 // COTAB/SotSB: copper .. jewelry; POD: platinum, gems, jewelry
	old   []byte
	multi *byte
	train *byte // POD only
}

func commonOf(r any) common {
	switch r := r.(type) {
	case *Record422:
		return common{&r.PName, &r.Stats, &r.Str00, r.SpellSlots[:], r.Coins[:], r.OldLevels[:], &r.MulticlassLevel, nil}
	case *Record439:
		return common{&r.PName, &r.Stats, &r.Str00, r.SpellSlots[:], r.Coins[:], r.OldLevels[:], &r.MulticlassLevel, nil}
	case *Record510:
		return common{&r.PName, &r.Stats, &r.Str00, r.SpellSlots[:], r.Coins[:], r.OldLevels[:], &r.MulticlassLevel, &r.CanTrain}
	}
	panic("char: not a later layout")
}

func sheetOfLater[R Later](c *rules1e.CharacterJSON, r R) {
	copySame(reflect.ValueOf(&c.Sheet).Elem(), reflect.ValueOf(r))
	m := commonOf(&r)
	n := min(int(m.name[0]), 15)
	c.Sheet.CharName = string(m.name[1 : 1+n])
	for i, p := range m.stats {
		c.Sheet.Stats[i] = rules1e.Score(p.Value)
	}
	c.Sheet.Str00 = m.str00.Value
	for cls, run := range slotRuns(&r) {
		fromBytes(c.Sheet.SpellSlots[cls][:run.n], m.slots[run.at:run.at+run.n])
	}
	copy(c.Sheet.Coins[len(c.Sheet.Coins)-len(m.coins):], m.coins)
	fromBytes(c.OldLevels[:], m.old)
	c.MulticlassLevel = rules1e.Level(*m.multi)
	if m.train != nil {
		c.CanTrain = *m.train != 0
	}
	if r, ok := any(r).(Record422); ok {
		c.Sheet.Pad86, c.Sheet.Pad87 = r.Pad86[0], r.Pad86[1]
	}
}

func laterOfSheet[R Later](r *R, c rules1e.CharacterJSON) {
	v := reflect.ValueOf(r).Elem()
	copySame(v, reflect.ValueOf(c.Sheet))
	m := commonOf(r)
	if n := min(int(m.name[0]), 15); string(m.name[1:1+n]) != c.Sheet.CharName { // a new name: the stale letters after it go
		*m.name = [16]byte{}
		m.name[0] = byte(copy(m.name[1:], c.Sheet.CharName[:min(len(c.Sheet.CharName), 15)]))
	}
	for i := range m.stats {
		m.stats[i].Value = byte(c.Sheet.Stats[i])
	}
	m.str00.Value = c.Sheet.Str00
	for cls, run := range slotRuns(r) {
		bytesOf(m.slots[run.at:run.at+run.n], c.Sheet.SpellSlots[cls][:run.n])
	}
	copy(m.coins, c.Sheet.Coins[len(c.Sheet.Coins)-len(m.coins):])
	bytesOf(m.old, c.OldLevels[:])
	*m.multi = byte(c.MulticlassLevel)
	if m.train != nil && (*m.train != 0) != c.CanTrain {
		*m.train = 0
		if c.CanTrain {
			*m.train = 1
		}
	}
	if p, ok := any(r).(*Record422); ok {
		p.Pad86 = [2]byte{c.Sheet.Pad86, c.Sheet.Pad87}
	}
}
