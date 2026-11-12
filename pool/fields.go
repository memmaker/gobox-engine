package pool

import (
	"errors"
	"fmt"
	"io/fs"
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
