package script

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
