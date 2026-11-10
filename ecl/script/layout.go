package script

import (
	"bufio"
	"fmt"
	"io"
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
}

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
	WidthText // a zero-terminated string, one character per address
)

var widthNames = map[string]Width{"byte": WidthByte, "word": WidthWord, "text": WidthText}

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
	// Replaces: an area's field that takes the place of the game's field on
	// exactly its addresses, for that area only: the area's script used the
	// game's word for state of its own (one script's reuse of another's word).
	Replaces bool
}

// Ref is a resolved operand: element Index of Field. A value: Args compare
// with ==, Ref and all; the zero Ref is none.
type Ref struct {
	Field FieldID
	Index int
}

// routine: f names an engine routine CALL runs, not memory. Routine numbers
// are their own space: they may share numbers with memory addresses.
func (f Field) routine() bool { return f.Owner == OwnerCall }

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
		return Ref{}, fmt.Errorf("routine %04X is no call field of the layout", a)
	}
	return Ref{}, fmt.Errorf("address %04X is in no field of the layout", a)
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
	m := &Layout{Fields: append(base, over.Fields...)}
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
		if i > 0 {
			if p := l.Fields[i-1]; p.routine() == f.routine() && p.From+p.Len > f.From {
				return fmt.Errorf("layout: %s overlaps %s", p.Name, f.Name)
			}
		}
	}
	return nil
}

// ParseLayout reads a layout text: one Field a line, "from len width owner
// name [readonly|replaces]", from in hex; '#' starts a comment.
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
		if len(f) != 5 && !(len(f) == 6 && (f[5] == "readonly" || f[5] == "replaces")) {
			return nil, fmt.Errorf("layout line %d: want from len width owner name [readonly|replaces]", n)
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
		if names[f[4]] {
			return nil, fmt.Errorf("layout line %d: %s twice", n, f[4])
		}
		names[f[4]] = true
		l.Fields = append(l.Fields, Field{Name: f[4], Owner: o, Width: w, From: int(from), Len: size, ReadOnly: len(f) == 6 && f[5] == "readonly", Replaces: len(f) == 6 && f[5] == "replaces"})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return l, l.check()
}
