package script

import (
	"bufio"
	"fmt"
	"io"
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
	Fields []Field // ordered by From; none overlap
}

// FieldID is a Field's place in its Layout; a Ref names its Field by it.
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
}

// Ref is a resolved operand: element Index of Field.
type Ref struct {
	Field FieldID
	Index int
}

// Resolve finds the Field holding address a.
func (l *Layout) Resolve(a int) (Ref, error) {
	i, _ := slices.BinarySearchFunc(l.Fields, a, func(f Field, a int) int { return f.From + f.Len - 1 - a })
	if i < len(l.Fields) && a >= l.Fields[i].From {
		return Ref{Field: FieldID(i), Index: a - l.Fields[i].From}, nil
	}
	return Ref{}, fmt.Errorf("address %04X is in no field of the layout", a)
}

// Field is the Field r names.
func (l *Layout) Field(r Ref) *Field { return &l.Fields[r.Field] }

// With is l and over together: an area's layout over the game's. A field of
// one may not overlap a field of the other, nor share its name.
func (l *Layout) With(over *Layout) (*Layout, error) {
	m := &Layout{Fields: append(slices.Clone(l.Fields), over.Fields...)}
	return m, m.check()
}

// check sorts the fields and refuses overlaps and repeated names.
func (l *Layout) check() error {
	slices.SortFunc(l.Fields, func(a, b Field) int { return a.From - b.From })
	names := map[string]bool{}
	for i, f := range l.Fields {
		if names[f.Name] {
			return fmt.Errorf("layout: %s twice", f.Name)
		}
		names[f.Name] = true
		if i > 0 {
			if p := l.Fields[i-1]; p.From+p.Len > f.From {
				return fmt.Errorf("layout: %s overlaps %s", p.Name, f.Name)
			}
		}
	}
	return nil
}

// ParseLayout reads a layout text: one Field a line, "from len width owner
// name [readonly]", from in hex; '#' starts a comment.
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
		if len(f) != 5 && !(len(f) == 6 && f[5] == "readonly") {
			return nil, fmt.Errorf("layout line %d: want from len width owner name [readonly]", n)
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
		l.Fields = append(l.Fields, Field{Name: f[4], Owner: o, Width: w, From: int(from), Len: size, ReadOnly: len(f) == 6})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return l, l.check()
}
