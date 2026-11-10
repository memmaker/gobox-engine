package script

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

// isVarArg: operand type t names state by its address.
func isVarArg(t byte) bool {
	return t == ArgVar || t == ArgVarWord || t == ArgVarLong || t == ArgTextVar
}

// ConstLayout is the layout of s's own tables: one read-only constant field
// a table, from its address to the next table's (or the end of the data).
func ConstLayout(s *Script, v *Version) *Layout {
	l := &Layout{}
	names := slices.Sorted(maps.Keys(s.Tables))
	slices.SortFunc(names, func(a, b string) int { return s.Tables[a] - s.Tables[b] })
	for i, n := range names {
		end := v.Base + len(s.Data)
		if i+1 < len(names) {
			end = s.Tables[names[i+1]]
		}
		if end > s.Tables[n] {
			l.Fields = append(l.Fields, Field{Name: n, Owner: OwnerConst, Width: WidthByte, From: s.Tables[n], Len: end - s.Tables[n], ReadOnly: true})
		}
	}
	return l
}

// ResolveRefs gives every variable operand of s its typed reference in l
// (the game's and the area's fields with ConstLayout's): an address no field
// covers is an error, naming the instruction; all of them are reported.
func ResolveRefs(s *Script, l *Layout) error {
	var errs []error
	for i := range s.Code {
		in := &s.Code[i]
		for k := range in.Args {
			a := &in.Args[k]
			if a.Label || !isVarArg(a.Type) {
				continue
			}
			r, err := l.Resolve(a.N)
			if err != nil {
				errs = append(errs, fmt.Errorf("area %d instruction %d (%s) operand %d: %w", s.Area, i, OpName(in.Op), k, err))
				continue
			}
			a.Ref = r
		}
	}
	return errors.Join(errs...)
}
