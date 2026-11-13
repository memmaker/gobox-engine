package rules1e

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gobox/rules"
)

// rosterFiles names a layout's character files by extension: the record's, then the
// files beside it (a missing one is none).
var rosterFiles = map[rules.RecordLayout]struct {
	LaterSave
	itemSize int
}{
	rules.Layout422: {LaterSave{Record: ".GUY", Items: ".SWG", Affects: ".FX"}, ItemSize},
	rules.Layout439: {LaterSave{Record: ".WHO", Items: ".STF", Affects: ".SFX", IDs: ".IDS"}, 67},
	rules.Layout510: {LaterSave{Record: ".PC", Items: ".THG", Affects: ".EFX"}, ItemSize},
}

// ReadRoster is the characters saved in dir by the game of record layout l, each as its
// canonical file (the one format every game's characters go through), by file name. A game's
// save files (.SAV) are not characters.
func ReadRoster(l rules.RecordLayout, dir string) ([]rules.CharacterFile, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var bases []string
	ext := ".cha"
	if f, ok := rosterFiles[l]; ok {
		ext = f.Record
	} else if l != rules.Layout285 {
		return nil, fmt.Errorf("ReadRoster: no character files known of layout %d", l)
	}
	for _, e := range ents {
		if n := e.Name(); !e.IsDir() && strings.EqualFold(filepath.Ext(n), ext) {
			bases = append(bases, strings.TrimSuffix(filepath.Join(dir, n), filepath.Ext(n)))
		}
	}
	slices.Sort(bases)
	var out []rules.CharacterFile
	for _, base := range bases {
		var c rules.CharacterFile
		if l == rules.Layout285 {
			c, err = Pack285(Files{Record: readFold(base, ext), Items: readFold(base, ".itm"), Affects: readFold(base, ".spc"),
				Plan: readFold(base, ".mem"), Dual: readFold(base, ".dual"), Powers: readFold(base, ".pow"),
				Portrait: readFold(base, ".png"), Icon: readFold(base, ".icon.png")})
		} else {
			c, err = packRoster(l, base)
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(base), err)
		}
		out = append(out, c)
	}
	return out, nil
}

// packRoster is the character at base of a later layout l.
func packRoster(l rules.RecordLayout, base string) (rules.CharacterFile, error) {
	s := rosterFiles[l]
	f := LaterFiles{Record: readFold(base, s.Record), Items: readFold(base, s.Items), Affects: readFold(base, s.Affects), ItemSize: s.itemSize}
	if b := readFold(base, s.IDs); s.IDs != "" && b != nil {
		f.Record = append(f.Record, PoolTrailer(ReadIDsFile(b))...)
	}
	switch l {
	case rules.Layout422:
		return PackLater[Record422](f)
	case rules.Layout439:
		return PackLater[Record439](f)
	}
	return PackLater[Record510](f)
}
