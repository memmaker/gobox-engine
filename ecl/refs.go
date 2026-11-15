package ecl

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"
)

// isVarArg: operand type t names state by its address.
func isVarArg(t byte) bool {
	return t == ArgVar || t == ArgVarWord || t == ArgVarLong || t == ArgTextVar
}

// tableBase: operand k of op is the base a table command indexes, whatever
// its operand type.
func tableBase(op byte, k int) bool {
	return op == OpGetTable && k == 0 || op == OpSaveTable && k == 1
}

// ConstLayout is the layout of s's own tables: one read-only constant field
// a table, from its address to the next table's (or the end of the data).
// GetTable reads the code byte at base plus index, unbounded, and the
// scripts lay tables over one another; a field is every entry a read of it
// can take, so no read leaves its field:
//   - a table read by a variable of a declared range (l's: Facing, range=4)
//     has that many entries: where the text's is shorter it runs on into the
//     tables after it (POR area 29: three exits-by-facing tables 3 bytes
//     apart, each read by four facings);
//   - a one-entry table a variable indexes is another base of the table
//     after it (cotab area21: T_8CFF[level] and T_8D00[level], one table
//     read 1-based and 0-based).
//
// The tables a field takes in are its named elements (their bases).
func ConstLayout(s *Script, v *Version, l *Layout) *Layout {
	out := &Layout{}
	names := slices.Sorted(maps.Keys(s.Tables))
	slices.SortFunc(names, func(a, b string) int { return s.Tables[a] - s.Tables[b] })
	reads := tableReads(s, l)
	runsOn := false // the last field has a read past its end: the next table continues it
	for i, n := range names {
		at, end := s.Tables[n], v.Base+len(s.Data)
		if i+1 < len(names) {
			end = s.Tables[names[i+1]]
		}
		if end <= at {
			continue
		}
		if runsOn {
			f := &out.Fields[len(out.Fields)-1]
			if f.Elements == nil {
				f.Elements = map[int]string{}
			}
			f.Elements[at-f.From] = n
			f.Len = end - f.From
		} else {
			out.Fields = append(out.Fields, Field{Name: n, Owner: OwnerConst, Width: WidthByte, From: at, Len: end - at, ReadOnly: true})
		}
		f := &out.Fields[len(out.Fields)-1]
		runsOn = false
		for base, r := range reads { // any base of the field read past the field's end
			if base >= f.From && base < f.From+f.Len && i+1 < len(names) {
				runsOn = runsOn || base+r.entries > f.From+f.Len || r.variable && end-at == 1 && base == at
			}
		}
	}
	return out
}

// tableRead is how s reads the table at a base address: by a variable, and
// the entries that variable's declared range takes (0: not declared).
type tableRead struct {
	variable bool
	entries  int
}

// tableReads are the table base addresses of s a variable indexes, with the
// widest declared range among the variables that do (l: the game's and the
// area's fields; nil: none declared).
func tableReads(s *Script, l *Layout) map[int]tableRead {
	t := map[int]tableRead{}
	for _, in := range s.Code {
		for k, a := range in.Args {
			if !tableBase(in.Op, k) || k+1 >= len(in.Args) || !isVarArg(in.Args[k+1].Type) {
				continue
			}
			r := t[a.N]
			r.variable = true
			if l != nil {
				if x, err := l.Resolve(in.Args[k+1].N); err == nil {
					r.entries = max(r.entries, l.Field(x).Range)
				}
			}
			t[a.N] = r
		}
	}
	return t
}

// ResolveRefs gives every variable operand of s its typed reference in l
// (the game's and the area's fields with ConstLayout's): an address no field
// covers is an error, naming the instruction; all of them are reported.
func ResolveRefs(s *Script, l *Layout) error {
	s.Layout = l
	var errs []error
	for i := range s.Code {
		in := &s.Code[i]
		for k := range in.Args {
			a := &in.Args[k]
			if a.Label || !(isVarArg(a.Type) || tableBase(in.Op, k)) {
				continue
			}
			resolve := l.Resolve
			if in.Op == OpCall {
				resolve = l.ResolveCall // a routine number, not memory
			}
			r, err := resolve(a.N)
			if l.Partial && errors.Is(err, errNoField) {
				continue // a partial layout leaves it a raw address
			}
			if err != nil {
				errs = append(errs, fmt.Errorf("area %d instruction %d (%s) operand %d: %w", s.Area, i, OpName(in.Op), k, err))
				continue
			}
			a.Ref = r
		}
	}
	return errors.Join(errs...)
}

// GlobalLayout and PartialLayout are the game's layout files in its script
// data: a complete one, or a Partial one (ReadLayout).
const (
	GlobalLayout  = "layout/global.txt"
	PartialLayout = "layout/partial.txt"
)

// ReadLayout is area's layout in a game's script data: layout/global.txt
// with layout/area<N>.txt (if the area has one) over it. Buck Rogers 1 and
// 2, whose own VM reads addresses, have layout/partial.txt instead: a
// Partial layout naming only the addresses the engine reads by name (until
// their full layouts are written, ARCHITECTURE.md reference queue 1b-i).
// Data with neither has none (nil, nil); every other game's VM refuses an
// unresolved operand.
func ReadLayout(data fs.FS, area int) (*Layout, error) {
	read := func(path string) (*Layout, error) {
		f, err := data.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		l, err := ParseLayout(f)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		return l, nil
	}
	g, err := read(GlobalLayout)
	if errors.Is(err, fs.ErrNotExist) {
		if g, err = read(PartialLayout); err == nil {
			g.Partial = true
		}
	}
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a, err := read(fmt.Sprintf("layout/area%d.txt", area))
	if errors.Is(err, fs.ErrNotExist) {
		return g, nil
	}
	if err != nil {
		return nil, err
	}
	return g.With(a)
}

// resolve gives s its Refs in its area's layout of data (none without one).
func resolve(s *Script, data fs.FS, v *Version) error {
	l, err := ReadLayout(data, s.Area)
	if err != nil || l == nil {
		return err
	}
	full, err := l.With(ConstLayout(s, v, l))
	if err != nil {
		return err
	}
	return ResolveRefs(s, full)
}
