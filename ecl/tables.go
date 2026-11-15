package ecl

import (
	"fmt"
	"io/fs"
)

// ScalarTables lists the commands of the imported texts in dir that index a
// field of one element (the bug class: a table the layout split into scalar
// names, indexed at run time; declare it one field of its full length, its
// entries element names). Each game's TestNoScalarTables runs it on its import.
func ScalarTables(v *Version, dir string, data fs.FS, fixes []Fix) ([]string, error) {
	var bad []string
	err := eachImportedArea(v, dir, data, fixes, func(area int, s *Script) {
		for i, in := range s.Code {
			for k, a := range in.Args {
				if !tableBase(in.Op, k) || a.Ref == (Ref{}) {
					continue
				}
				f, idx := s.Layout.Field(a.Ref), in.Args[k+1]
				switch {
				case idx.Ref != (Ref{}) && f.Len == 1:
					bad = append(bad, fmt.Sprintf("area %d command %d (%s): indexes %s, a field of one element, by %s", area, i, OpName(in.Op), f.Name, s.Layout.Field(idx.Ref).Name))
				case idx.Ref == (Ref{}) && a.Ref.Index+idx.N >= f.Len:
					bad = append(bad, fmt.Sprintf("area %d command %d (%s): %s[%d] is past the field's %d", area, i, OpName(in.Op), f.Name, idx.N, f.Len))
				}
			}
		}
	})
	return bad, err
}

// TableRead is a command of an imported text that reads one of the script's
// own tables (a constant field) by a variable: the bound of the index is
// the variable's range, not the text's.
type TableRead struct {
	Area   int
	Table  string // the field read, its Len entries
	Len    int
	Offset int    // the base's place in the field (a later base of a field of several)
	Index  string // the variable that indexes it
	Range  int    // the values the variable takes, 0..Range-1 (0: not declared)
}

// TableReads lists the TableRead of the imported texts in dir, each once an area.
func TableReads(v *Version, dir string, data fs.FS, fixes []Fix) ([]TableRead, error) {
	var out []TableRead
	err := eachImportedArea(v, dir, data, fixes, func(area int, s *Script) {
		seen := map[TableRead]bool{}
		for _, in := range s.Code {
			for k, a := range in.Args {
				if !tableBase(in.Op, k) || a.Ref == (Ref{}) || k+1 >= len(in.Args) || in.Args[k+1].Ref == (Ref{}) {
					continue
				}
				f, x := s.Layout.Field(a.Ref), s.Layout.Field(in.Args[k+1].Ref)
				if f.Owner != OwnerConst {
					continue
				}
				r := TableRead{Area: area, Table: f.Name, Len: f.Len, Offset: a.Ref.Index, Index: x.Name, Range: x.Range}
				if !seen[r] {
					seen[r] = true
					out = append(out, r)
				}
			}
		}
	})
	return out, err
}
