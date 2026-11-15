package ecl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// A Layout is a game's map from the byte code's addresses to typed state:
// every address a script may name belongs to exactly one Field, and the
// importer turns each operand into a Ref to it. An address no Field covers
// fails the import. Each game supplies its layout as data, written from the
// record types of its Pascal (its -re repo).
type Layout struct {
	Fields []Field // memory fields by From, then routine fields by From; none overlap within its space
	// Partial: the layout names only some addresses (Buck Rogers 1 and 2's,
	// layout/partial.txt): an operand no Field covers stays a raw address
	// (V_XXXX) instead of failing the import.
	Partial bool
}

// errNoField is find's error for an address no Field covers.
var errNoField = errors.New("in no field of the layout")

// FieldID is a Field's place in its Layout, from 1; a Ref names its Field by
// it, and the zero FieldID is no field (an operand not resolved).
type FieldID int

// Owner is whose state a Field is.
type Owner uint8

const (
	OwnerArea   Owner = iota + 1 // the current area's saved state
	OwnerGame                    // the game's saved state, across areas
	OwnerMember                  // the selected party member's record
	OwnerLocal                   // the script's scratch, cleared by a fresh area
	OwnerConst                   // a constant table in the script's own code
	OwnerFact                    // an engine fact (position, facing, the wall ahead)
	OwnerCall                    // an engine routine CALL names
)

var ownerNames = map[string]Owner{"area": OwnerArea, "game": OwnerGame, "member": OwnerMember,
	"local": OwnerLocal, "const": OwnerConst, "fact": OwnerFact, "call": OwnerCall}

// Width is how wide one element of a Field is.
type Width uint8

const (
	WidthByte Width = iota + 1
	WidthWord
	WidthText  // a zero-terminated string, one character per address
	WidthDword // four bytes, one address (the original's longints)
	// WidthBytes: Len raw bytes, one address each, that operands of any
	// width read and write little-endian at an offset (Buck Rogers' scratch
	// zones 4C00..4C1F and 7F79..7F82). Only a dialect that stores by operand has them.
	WidthBytes
)

var widthNames = map[string]Width{"byte": WidthByte, "word": WidthWord, "dword": WidthDword, "text": WidthText, "bytes": WidthBytes}

// Field is one typed piece of state: Len elements of Width, one address
// each, from address From on. Len 1 is a single value; more is an array a
// computed address (base plus index) indexes.
type Field struct {
	Name     string // its name in the script text and in saves
	Owner    Owner
	Width    Width
	From     int
	Len      int
	ReadOnly bool // the byte code only reads it (an engine fact, a constant table)
	// Range: the values the field takes are 0..Range-1 (range=N; 0: not
	// declared). A table of the script's own that it indexes has that many
	// entries: where the text's is shorter it runs on into the next (ConstLayout).
	Range int
	// Replaces: an area's field that takes the place of the game's field on
	// exactly its addresses, for that area only: the area's script used the
	// game's word for state of its own (one script's reuse of another's word).
	Replaces bool
	// Encoding: the game keeps this canonical field in an encoding of its own (a value
	// or element numbering other than POR's); the engine turns it into the canonical
	// value by this name. "": the canonical encoding.
	Encoding Encoding
	// Coins: the field's elements are money (coins=, CoinCell an element); nil: none.
	Coins []CoinCell
	// Elements: names of single elements of an array, by index (the layout's
	// "from = name" lines): a table the scripts index, whose entries other
	// scripts name one by one (POR's QuestTable, QuestSokalKeep its element 1).
	Elements map[int]string
}

// ElementName is the name of f's element i, or "": the element has none.
func (f *Field) ElementName(i int) string { return f.Elements[i] }

// Encoding names a game's encoding of a canonical field (Field.Encoding).
type Encoding string

// Ref is a resolved operand: element Index of Field. A value: Args compare
// with ==, Ref and all; the zero Ref is none. Two Refs name the same cell
// when their Cell()s are equal.
type Ref struct {
	Field FieldID
	Index int
	// As is the width of the operand the script addresses the cell through
	// (a byte, word or long variable); zero: the engine addresses the field
	// itself, at the field's own width.
	As Width
}

// Cell is r without the operand's width: the cell alone, for "the same cell" tests.
func (r Ref) Cell() Ref { r.As = 0; return r }

// routine: f names an engine routine CALL runs, not memory. Routine numbers
// are their own space: they may share numbers with memory addresses.
func (f Field) routine() bool { return f.Owner == OwnerCall }

// Plus is the element i past r in its field's memory: a table operand's
// base and the index a command adds (GET TABLE, SAVE TABLE). Like the byte
// code's address arithmetic it may pass the field's end.
func (r Ref) Plus(i int) Ref { r.Index += i; return r }

// Named is the memory field called name: for the words an engine reads or
// writes by itself (a native command's answer, DIVIDE's remainder).
func (l *Layout) Named(name string) (Ref, error) {
	for i, f := range l.Fields {
		if f.routine() {
			continue
		}
		if f.Name == name {
			return Ref{Field: FieldID(i + 1)}, nil
		}
		for k, n := range f.Elements {
			if n == name {
				return Ref{Field: FieldID(i + 1), Index: k}, nil
			}
		}
	}
	return Ref{}, fmt.Errorf("layout: no field %s", name)
}

// Address is r's address in the byte code's memory: for a game whose state
// is still that memory (its host maps a Ref back to it).
func (l *Layout) Address(r Ref) int { f := l.Field(r); return f.From + r.Index }

// Resolve finds the memory Field holding address a.
func (l *Layout) Resolve(a int) (Ref, error) { return l.find(a, false) }

// ResolveCall finds the routine Field CALL's operand a names.
func (l *Layout) ResolveCall(a int) (Ref, error) { return l.find(a, true) }

func (l *Layout) find(a int, routine bool) (Ref, error) {
	for i, f := range l.Fields {
		if f.routine() == routine && a >= f.From && a < f.From+f.Len {
			return Ref{Field: FieldID(i + 1), Index: a - f.From}, nil
		}
	}
	if routine {
		return Ref{}, fmt.Errorf("routine %04X: %w", a, errNoField)
	}
	return Ref{}, fmt.Errorf("address %04X: %w", a, errNoField)
}

// Field is the Field r names.
func (l *Layout) Field(r Ref) *Field { return &l.Fields[r.Field-1] }

// With is l and over together: an area's layout over the game's. A field of
// one may not overlap a field of the other, nor share its name, except that
// an over field marked Replaces takes the place of the l field with exactly
// its addresses (and must find one).
func (l *Layout) With(over *Layout) (*Layout, error) {
	base := slices.Clone(l.Fields)
	for _, o := range over.Fields {
		if !o.Replaces {
			continue
		}
		i := slices.IndexFunc(base, func(f Field) bool { return f.From == o.From && f.Len == o.Len && f.routine() == o.routine() })
		if i < 0 {
			return nil, fmt.Errorf("layout: %s replaces no field at %04X", o.Name, o.From)
		}
		base = slices.Delete(base, i, i+1)
	}
	m := &Layout{Fields: append(base, over.Fields...), Partial: l.Partial || over.Partial}
	return m, m.check()
}

// elementName is the form ReadNames gives an array's elements (Name_i).
var elementName = regexp.MustCompile(`_[0-9]+$`)

// check sorts the fields (memory, then routines, each by address) and
// refuses overlaps within a space and repeated names.
func (l *Layout) check() error {
	slices.SortFunc(l.Fields, func(a, b Field) int {
		if a.routine() != b.routine() {
			if a.routine() {
				return 1
			}
			return -1
		}
		return a.From - b.From
	})
	names := map[string]bool{}
	for i, f := range l.Fields {
		if names[f.Name] {
			return fmt.Errorf("layout: %s twice", f.Name)
		}
		if f.Owner != OwnerConst && elementName.MatchString(f.Name) { // a table keeps the decoder's T_<addr>
			return fmt.Errorf("layout: %s ends in _<number>, the form of an array element's name", f.Name)
		}
		names[f.Name] = true
		for _, n := range f.Elements {
			if names[n] {
				return fmt.Errorf("layout: %s twice", n)
			}
			names[n] = true
		}
		if i > 0 {
			if p := l.Fields[i-1]; p.routine() == f.routine() && p.From+p.Len > f.From {
				return fmt.Errorf("layout: %s overlaps %s", p.Name, f.Name)
			}
		}
	}
	return nil
}

// ParseLayout reads a layout text: one Field a line, "from len width owner
// name [readonly|replaces] [enc=<encoding>] [coins=<cells>] [range=<n>]", from in hex, or
// "from = name": the name of the element at from of the array field before
// it (Field.Elements); '#' starts a comment.
func ParseLayout(r io.Reader) (*Layout, error) {
	l := &Layout{}
	names := map[string]bool{}
	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		line, _, _ := strings.Cut(sc.Text(), "#")
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if len(f) == 3 && f[1] == "=" {
			if err := l.nameElement(f[0], f[2]); err != nil {
				return nil, fmt.Errorf("layout line %d: %w", n, err)
			}
			continue
		}
		if len(f) < 5 {
			return nil, fmt.Errorf("layout line %d: want from len width owner name [readonly|replaces] [enc=<encoding>] [coins=<cells>]", n)
		}
		var readOnly, replaces bool
		var enc Encoding
		var coins string
		rng := 0
		for _, flag := range f[5:] {
			c, isCoins := strings.CutPrefix(flag, "coins=")
			if r, isRange := strings.CutPrefix(flag, "range="); isRange && rng == 0 {
				if rng, _ = strconv.Atoi(r); rng < 1 {
					return nil, fmt.Errorf("layout line %d: range %q", n, r)
				}
				continue
			}
			switch e, isEnc := strings.CutPrefix(flag, "enc="); {
			case isCoins && c != "" && coins == "":
				coins = c
			case flag == "readonly" && !readOnly:
				readOnly = true
			case flag == "replaces" && !replaces:
				replaces = true
			case isEnc && e != "" && enc == "":
				enc = Encoding(e)
			default:
				return nil, fmt.Errorf("layout line %d: flag %q", n, flag)
			}
		}
		from, err := strconv.ParseInt(f[0], 16, 32)
		if err != nil {
			return nil, fmt.Errorf("layout line %d: from %q", n, f[0])
		}
		size, err := strconv.Atoi(f[1])
		if err != nil || size < 1 {
			return nil, fmt.Errorf("layout line %d: len %q", n, f[1])
		}
		w, ok := widthNames[f[2]]
		if !ok {
			return nil, fmt.Errorf("layout line %d: width %q", n, f[2])
		}
		o, ok := ownerNames[f[3]]
		if !ok {
			return nil, fmt.Errorf("layout line %d: owner %q", n, f[3])
		}
		if w == WidthBytes && (o == OwnerConst || o == OwnerFact || o == OwnerCall) {
			return nil, fmt.Errorf("layout line %d: %s: raw bytes make no %s", n, f[4], f[3])
		}
		if names[f[4]] {
			return nil, fmt.Errorf("layout line %d: %s twice", n, f[4])
		}
		names[f[4]] = true
		fd := Field{Name: f[4], Owner: o, Width: w, From: int(from), Len: size, ReadOnly: readOnly, Replaces: replaces, Encoding: enc, Range: rng}
		if coins != "" {
			if fd.Coins, err = parseCoins(coins, size); err != nil {
				return nil, fmt.Errorf("layout line %d: %w", n, err)
			}
		}
		l.Fields = append(l.Fields, fd)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return l, l.check()
}

// nameElement names the element at address from (hex) of the last field read,
// which must be an array holding it.
func (l *Layout) nameElement(from, name string) error {
	a, err := strconv.ParseInt(from, 16, 32)
	if err != nil {
		return fmt.Errorf("from %q", from)
	}
	if len(l.Fields) == 0 {
		return fmt.Errorf("element %s before any field", name)
	}
	f := &l.Fields[len(l.Fields)-1]
	i := int(a) - f.From
	if f.Len < 2 || f.routine() || i < 0 || i >= f.Len {
		return fmt.Errorf("element %s at %04X is in no array field before it (%s)", name, a, f.Name)
	}
	if elementName.MatchString(name) {
		return fmt.Errorf("element %s ends in _<number>", name)
	}
	if f.Elements == nil {
		f.Elements = map[int]string{}
	}
	if _, dup := f.Elements[i]; dup {
		return fmt.Errorf("element %04X of %s named twice", a, f.Name)
	}
	f.Elements[i] = name
	return nil
}

// GameFields are the fields other than the locals of the layouts in
// data (global.txt and every area<N>.txt over it), each name once.
func GameFields(data fs.FS) ([]Field, error) {
	names, err := fs.Glob(data, "layout/area*.txt")
	if err != nil {
		return nil, err
	}
	areas := []int{0}
	for _, n := range names {
		a, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(n, "layout/area"), ".txt"))
		if err != nil {
			return nil, fmt.Errorf("layout file %s: %w", n, err)
		}
		areas = append(areas, a)
	}
	var out []Field
	seen := map[string]bool{}
	for _, a := range areas {
		l, err := ReadLayout(data, a)
		if err != nil {
			return nil, err
		}
		if l == nil {
			return nil, errors.New("no layout/global.txt")
		}
		for _, f := range l.Fields {
			if f.Owner != OwnerLocal && !seen[f.Name] {
				seen[f.Name] = true
				out = append(out, f)
			}
		}
	}
	return out, nil
}
