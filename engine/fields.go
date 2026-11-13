package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"gobox/ecl/script"
)

// The game's fields: those of every area's layout but the locals, by name.

// saved: field f keeps cells of its own (an original save's records carry
// them; saveimport.go).
func (e *Engine) saved(f script.Field) bool {
	switch f.Owner {
	case script.OwnerGame, script.OwnerArea, script.OwnerMember, script.OwnerLocal:
	default:
		return false
	}
	if ef, ok := e.engineFields[f.Name]; f.Replaces || ok && !ef.stored {
		return false
	}
	return true
}

// gameFields are the saved fields of every area's layout (the game's state
// is kept across areas), each once.
func (e *Engine) gameFields() []script.Field {
	if e.savedFields == nil {
		fields, err := layoutGameFields(os.DirFS(e.scriptDir()))
		if err != nil {
			panic(err) // the game's script layout is broken: a broken game folder
		}
		e.savedFields = slices.DeleteFunc(fields, func(f script.Field) bool { return !e.saved(f) })
	}
	return e.savedFields
}

// layoutGameFields are the fields other than the locals of the layouts in
// data (global.txt and every area<N>.txt over it), each name once.
func layoutGameFields(data fs.FS) ([]script.Field, error) {
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
	var out []script.Field
	seen := map[string]bool{}
	for _, a := range areas {
		l, err := script.ReadLayout(data, a)
		if err != nil {
			return nil, err
		}
		if l == nil {
			return nil, errors.New("no layout/global.txt")
		}
		for _, f := range l.Fields {
			if f.Owner != script.OwnerLocal && !seen[f.Name] {
				seen[f.Name] = true
				out = append(out, f)
			}
		}
	}
	return out, nil
}

// foldElements moves the cells a save keeps under a named element's own name
// into its array's cells: a save written while the layout had the element as
// a field of its own (POR's QuestSokalKeep before QuestTable, cotab's Local00
// before Locals) loads unchanged. A name some field of fields has stays.
func foldElements(c fieldCells, fields []script.Field) {
	own := map[string]bool{}
	for _, f := range fields {
		own[f.Name] = true
	}
	for _, f := range fields {
		for _, i := range slices.Sorted(maps.Keys(f.Elements)) {
			n := f.Elements[i]
			v, ok := c[n]
			if !ok || own[n] {
				continue
			}
			a := c[f.Name]
			if len(a) < i+len(v) {
				a = append(a, make([]int64, i+len(v)-len(a))...)
			}
			copy(a[i:], v)
			c[f.Name] = a
			delete(c, n)
		}
	}
}
