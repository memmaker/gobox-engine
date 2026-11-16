package engine

import (
	"maps"
	"os"
	"slices"

	"gobox/ecl"
)

// The game's fields: those of every area's layout but the locals, by name.

// saved: field f keeps cells of its own (an original save's records carry
// them; importer saveimport.go).
func (e *Engine) saved(f ecl.Field) bool {
	switch f.Owner {
	case ecl.OwnerGame, ecl.OwnerArea, ecl.OwnerMember, ecl.OwnerLocal:
	default:
		return false
	}
	if ef, ok := e.bound()[f.Name]; f.Replaces || ok && !ef.stored {
		return false
	}
	return true
}

// gameFields are the saved fields of every area's layout (the game's state
// is kept across areas), each once.
func (e *Engine) gameFields() []ecl.Field {
	if e.savedFields == nil {
		fields, err := ecl.GameFields(os.DirFS(e.scriptDir()))
		if err != nil {
			panic(err) // the game's script layout is broken: a broken game folder
		}
		e.savedFields = slices.DeleteFunc(fields, func(f ecl.Field) bool { return !e.saved(f) })
	}
	return e.savedFields
}

// foldElements moves the cells a save keeps under a named element's own name
// into its array's cells: a save written while the layout had the element as
// a field of its own (POR's QuestSokalKeep before QuestTable, cotab's Local00
// before Locals) loads unchanged. A name some field of fields has stays.
func foldElements(c fieldCells, fields []ecl.Field) {
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
