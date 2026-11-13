package rules1e

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"gobox/rules"
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
// the Sheet and the extras, the record itself kept for the rest. Memorized
// and SpellBook hold the game's spell numbers; rules1e's Import maps them
// to the unified ids.
func PackLater[R Later](f LaterFiles) (rules.CharacterFile, error) {
	n := len(Encode(*new(R)))
	rec, ids := SplitPoolIDs(f.Record, n)
	r := Decode[R](rec)
	c := CharacterJSON{PoolIDs: ids, Kept: &KeptRecord{Layout: rules.RecordLayout(n), Image: rules.RecordImage(bytes.Clone(rec))}}
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
func UnpackLater[R Later](b rules.CharacterFile, itemSize int) (LaterFiles, error) {
	var c CharacterJSON
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
	f := LaterFiles{Record: append(withBlanks[R](Encode(r), orig), PoolTrailer(c.PoolIDs)...), ItemSize: itemSize}
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
	"PName": true, "Stats": true, "Str00": true, "Memorized": true,
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
func slotRuns(r any) (runs [SpellClasses]struct{ at, n int }) {
	set := func(cleric, druid, mage, n int) {
		runs[SpellClassCleric] = struct{ at, n int }{cleric, n}
		runs[SpellClassDruid] = struct{ at, n int }{druid, n}
		runs[SpellClassMage] = struct{ at, n int }{mage, n}
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
	memo  []byte
	slots []byte
	coins []uint16 // COTAB/SotSB: copper .. jewelry; POD: platinum, gems, jewelry
	old   []byte
	multi *byte
	train *byte // POD only
}

func commonOf(r any) common {
	switch r := r.(type) {
	case *Record422:
		return common{&r.PName, &r.Stats, &r.Str00, r.Memorized[:], r.SpellSlots[:], r.Coins[:], r.OldLevels[:], &r.MulticlassLevel, nil}
	case *Record439:
		return common{&r.PName, &r.Stats, &r.Str00, r.Memorized[:], r.SpellSlots[:], r.Coins[:], r.OldLevels[:], &r.MulticlassLevel, nil}
	case *Record510:
		return common{&r.PName, &r.Stats, &r.Str00, r.Memorized[:], r.SpellSlots[:], r.Coins[:], r.OldLevels[:], &r.MulticlassLevel, &r.CanTrain}
	}
	panic("char: not a later layout")
}

func sheetOfLater[R Later](c *CharacterJSON, r R) {
	copySame(reflect.ValueOf(&c.Sheet).Elem(), reflect.ValueOf(r))
	m := commonOf(&r)
	n := min(int(m.name[0]), 15)
	c.Sheet.CharName = string(m.name[1 : 1+n])
	for i, p := range m.stats {
		c.Sheet.Stats[i], c.Sheet.Base[i] = Score(p.Value), Score(p.Base)
	}
	c.Sheet.Str00, c.Sheet.Str00Base = m.str00.Value, m.str00.Base
	memoryOf(c.Sheet.Memorized[:], m.memo)
	for cls, run := range slotRuns(&r) {
		fromBytes(c.Sheet.SpellSlots[cls][:run.n], m.slots[run.at:run.at+run.n])
	}
	copy(c.Sheet.Coins[len(c.Sheet.Coins)-len(m.coins):], m.coins)
	fromBytes(c.OldLevels[:], m.old)
	c.MulticlassLevel = Level(*m.multi)
	if m.train != nil {
		c.CanTrain = *m.train != 0
	}
	if r, ok := any(r).(Record422); ok {
		c.Sheet.ScriptFlag = r.Pad402[0]
		c.Sheet.Pad86, c.Sheet.Pad87 = r.Pad86[0], r.Pad86[1]
	}
}

func laterOfSheet[R Later](r *R, c CharacterJSON) {
	v := reflect.ValueOf(r).Elem()
	copySame(v, reflect.ValueOf(c.Sheet))
	m := commonOf(r)
	if n := min(int(m.name[0]), 15); string(m.name[1:1+n]) != c.Sheet.CharName { // a new name: the stale letters after it go
		*m.name = [16]byte{}
		m.name[0] = byte(copy(m.name[1:], c.Sheet.CharName[:min(len(c.Sheet.CharName), 15)]))
	}
	base, base00 := c.Sheet.Base, c.Sheet.Str00Base
	if c.Kept == nil && base == ([6]Score{}) { // a character made here: its own scores are those in use
		base, base00 = c.Sheet.Stats, c.Sheet.Str00
	}
	for i := range m.stats {
		m.stats[i].Value, m.stats[i].Base = byte(c.Sheet.Stats[i]), byte(base[i])
	}
	m.str00.Value, m.str00.Base = c.Sheet.Str00, base00
	memoryBytes(m.memo, c.Sheet.Memorized[:])
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
		p.Pad402[0] = c.Sheet.ScriptFlag
		p.Pad86 = [2]byte{c.Sheet.Pad86, c.Sheet.Pad87}
	}
}

// CheckLater is f through PackLater, rules1e's Import and Export over the
// game's tables tb and base item table items, and UnpackLater: nil when the
// record, items and affects come back byte-equal. Import fails on a spell,
// affect or item power number the game's maps (rules.json) lack.
func CheckLater[R Later](tb *Tables, items []ItemBase, f LaterFiles) error {
	b, err := PackLater[R](f)
	if err != nil {
		return err
	}
	r := &Rules{S: &RuleState{Tables: tb, Items: items}}
	id, err := r.Import(b)
	if err != nil {
		return err
	}
	if b, err = r.Export(id); err != nil {
		return err
	}
	g, err := UnpackLater[R](b, f.ItemSize)
	if err != nil {
		return err
	}
	// a port-written save's old trailer comes back as PoolTrailer: its ids
	// must, the record's bytes too
	n := len(Encode(*new(R)))
	wantRec, wantIDs := SplitPoolIDs(f.Record, n)
	gotRec, gotIDs := SplitPoolIDs(g.Record, n)
	if gotIDs != wantIDs {
		return fmt.Errorf("pool ids %+v, want %+v", gotIDs, wantIDs)
	}
	if len(f.Record) <= n || bytes.HasPrefix(f.Record[n:], []byte(PoolMagic)) {
		gotRec, wantRec = g.Record, f.Record // no trailer or the canonical one: byte for byte
	}
	if !bytes.Equal(gotRec, wantRec) {
		for i := range min(len(gotRec), len(wantRec)) {
			if gotRec[i] != wantRec[i] {
				return fmt.Errorf("record byte %#x: %d, want %d", i, gotRec[i], wantRec[i])
			}
		}
		return fmt.Errorf("record length %d, want %d", len(gotRec), len(wantRec))
	}
	// the DOS pointers in items and affects are not kept (the port writes zeros)
	pad := make([]byte, f.ItemSize-ItemSize)
	var want []byte
	kept := readItems(f.Items, f.ItemSize)
	if err := r.GoneSpellsOut(kept); err != nil {
		return err
	}
	for _, it := range kept {
		want = append(append(want, WriteItems([]Item{it})...), pad...)
	}
	if !bytes.Equal(g.Items, want) {
		return firstDiff("items", g.Items, want, f.ItemSize)
	}
	if affects := WriteAffects(ReadAffects(f.Affects)); !bytes.Equal(g.Affects, affects) {
		return firstDiff("affects", g.Affects, affects, AffectSize)
	}
	return nil
}

// firstDiff names the first n-byte record of got that differs from want.
func firstDiff(what string, got, want []byte, n int) error {
	for i := range min(len(got), len(want)) {
		if got[i] != want[i] {
			return fmt.Errorf("%s record %d byte %#x: %d, want %d", what, i/n, i%n, got[i], want[i])
		}
	}
	return fmt.Errorf("%s length %d, want %d", what, len(got), len(want))
}

// LaterSave names a game's character files by extension (".GUY", ".SWG",
// ".FX"); a missing items or affects file is none.
// IDs (SotSB's old .ids) is read as the ids its trailer now carries.
type LaterSave struct{ Record, Items, Affects, IDs string }

// CheckLaterDir runs CheckLater on every character file under dir (any of
// saves' record extensions, any case), their items held ItemSize-wide in
// itemSize bytes, and on one character (the first found) holding extra as
// its items: the game data's items, so every spell, power and affect number
// they hold meets the maps. It returns the files checked and their failures.
func CheckLaterDir[R Later](tb *Tables, items []ItemBase, dir string, itemSize int, extra []Item, saves ...LaterSave) (checked []string, errs []error) {
	var first []byte
	filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := filepath.Ext(p)
		for _, s := range saves {
			if !strings.EqualFold(ext, s.Record) {
				continue
			}
			base := strings.TrimSuffix(p, ext)
			f := LaterFiles{Record: readFold(base, s.Record), Items: readFold(base, s.Items), Affects: readFold(base, s.Affects), ItemSize: itemSize}
			if b := readFold(base, s.IDs); s.IDs != "" && b != nil {
				f.Record = append(f.Record, PoolTrailer(ReadIDsFile(b))...)
			}
			checked = append(checked, p)
			if first == nil {
				first = f.Record
			}
			if err := CheckLater[R](tb, items, f); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", p, err))
			}
		}
		return nil
	})
	if first != nil && len(extra) > 0 {
		pad := make([]byte, itemSize-ItemSize)
		f := LaterFiles{Record: first, ItemSize: itemSize}
		for _, it := range extra {
			it.Readied = 0 // loot: one character cannot ready them all
			f.Items = append(append(f.Items, WriteItems([]Item{it})...), pad...)
		}
		for i := 0; i < len(extra); i += 100 { // in hundreds: a failure names its range
			g := f
			g.Items = f.Items[i*itemSize : min(i+100, len(extra))*itemSize]
			if err := CheckLater[R](tb, items, g); err != nil {
				errs = append(errs, fmt.Errorf("data items %d..: %w", i, err))
			}
		}
	}
	return checked, errs
}

// readFold reads base+ext in the case ext has or upper/lower case; nil: none.
func readFold(base, ext string) []byte {
	for _, e := range []string{ext, strings.ToUpper(ext), strings.ToLower(ext)} {
		if b, err := os.ReadFile(base + e); err == nil {
			return b
		}
	}
	return nil
}
