package ecl

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Library is a game's imported scripts: one text per area in Dir, written
// by Import. It only reads them: a missing or stale text is ErrNotImported.
type Library struct {
	Version *Version
	Dir     string  // where Import wrote the texts
	Data    fs.FS   // the game's script data: names/, layout/
	Fixes   []Fix   // the game's decompiler special cases, part of the key
	Fold    Folding // FoldGates, or KeepGates for a test comparing the two

	key string // Key, once
}

// ErrNotImported: the scripts are missing or were imported from other data
// or by another decompiler; the import must run.
var ErrNotImported = errors.New("scripts not imported or stale: run the import")

// dataDirs are the parts of a game's script data a text is made from.
var dataDirs = []string{"names", "layout"}

// Key is what the texts are made from: the decompiler's version, the
// folding, the game's script data (dataDirs: every file, path and content)
// and its special cases. A text whose header carries another key is stale.
func Key(v *Version, data fs.FS, fold Folding, fixes []Fix) (string, error) {
	h := sha256.New()
	fmt.Fprintf(h, "%s\x00fold %d\x00", v.Header(), fold)
	for _, f := range fixes {
		fmt.Fprintf(h, "fix %d\x00%d\x00%s\x00%d\x00%s\x00", f.Area, len(f.Old), f.Old, len(f.New), f.New)
	}
	err := fs.WalkDir(data, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != "." && !slices.Contains(dataDirs, strings.SplitN(p, "/", 2)[0]) {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.Contains(p, "/") {
			return nil // a file beside the data folders: not data (POR's texts share its folder)
		}
		b, err := fs.ReadFile(data, p)
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "%s\x00%d\x00", p, len(b))
		h.Write(b)
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Empty is the script of an area with no block: every event exits.
func Empty(area int) *Script {
	return &Script{Area: area, Code: []Instr{{Op: OpExit}}, Labels: map[string]int{}, At: map[int]string{}, Tables: map[string]int{}}
}

// Area is area's script, Empty if the import wrote none (the game has no
// such block).
func (l *Library) Area(area int) (*Script, error) {
	names, err := ReadNames(l.Data, area, false, l.Version)
	if err != nil {
		return nil, err
	}
	if l.key == "" {
		if l.key, err = Key(l.Version, l.Data, l.Fold, l.Fixes); err != nil {
			return nil, err
		}
	}
	path := AreaPath(l.Dir, area)
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		any, err := filepath.Glob(filepath.Join(l.Dir, "area*.txt"))
		if err != nil {
			return nil, err
		}
		if len(any) == 0 {
			return nil, fmt.Errorf("%s: %w", l.Dir, ErrNotImported)
		}
		return Empty(area), nil
	} else if err != nil {
		return nil, err
	}
	if _, err := l.Version.checkHeader(string(b)); err != nil || textKey(string(b)) != l.key {
		return nil, fmt.Errorf("%s: %w", path, ErrNotImported)
	}
	s, err := Parse(string(b), names)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", path, err)
	}
	s.Area = area
	if err := checkOperands(s, l.Data); err != nil {
		return nil, fmt.Errorf("%s:%w", path, err)
	}
	if err := resolve(s, l.Data, l.Version); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return s, nil
}

// writeAtomic writes path so that no reader sees a part of it.
func writeAtomic(path, text string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "import*")
	if err != nil {
		return err
	}
	_, err = f.WriteString(text)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		return errors.Join(err, os.Remove(f.Name()))
	}
	return nil
}

// CacheDir is where a game with no scripts folder of its own keeps its
// imported texts: they are made from the game's data, which may be read-only.
// It is created when missing (a first run, a fresh home).
func CacheDir(game string) (string, error) {
	d, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(d, "gobox", "scripts", game)
	return dir, os.MkdirAll(dir, 0o755)
}


// Assemble makes a script of raw byte code with no header, one command per
// command, for a test to run a few commands: command numbers are its jumps;
// its operands resolve in lay (none for nil: Buck Rogers, ReadLayout).
func Assemble(v *Version, lay *Layout, code []byte) (*Script, error) {
	s := Empty(0)
	s.Code = nil
	r := &reader{code: code, base: v.Base}
	for pc := v.Base; pc < v.Base+len(code); {
		l := v.decode(r, pc)
		s.Code = append(s.Code, Instr{Op: l.Cmd.Op, Args: l.Args})
		pc = l.Next
	}
	if lay == nil {
		return s, nil
	}
	return s, ResolveRefs(s, lay)
}
