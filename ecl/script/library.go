package script

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

	"gobox/dax"
)

// Library is a game's decompiled scripts: one text per area in Dir, each
// made from the game's data the first time it is needed or when the
// decompiler has moved on.
type Library struct {
	Version *Version
	Source  Source
	Dir     string // where the texts are written
	Data    fs.FS  // the game's script data: names/, layout/, patches/
}

// dataDirs are the parts of a game's script data a text is made from.
var dataDirs = []string{"names", "layout", "patches"}

// key is what an area's text is made from: the decompiler's version, the
// game's script data (dataDirs: every file, path and content) and the
// area's code. A text whose header carries another key is stale.
func (l *Library) key(code []byte) (string, error) {
	h := sha256.New()
	fmt.Fprintf(h, "%s\x00", l.Version.Header())
	err := fs.WalkDir(l.Data, ".", func(p string, d fs.DirEntry, err error) error {
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
		b, err := fs.ReadFile(l.Data, p)
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
	fmt.Fprintf(h, "code\x00%d\x00", len(code))
	h.Write(code)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// DAXSource reads the blocks of the game's ECL*.DAX.
func DAXSource(l *dax.Loader) Source {
	return Source{Block: func(disk, block int) []byte {
		name := fmt.Sprintf("ECL%d", disk)
		if !l.Exists(name + ".DAX") {
			return nil
		}
		return l.BlockShared(name, block)
	}}
}

// Empty is the script of an area with no block: every event exits.
func Empty(area int) *Script {
	return &Script{Area: area, Code: []Instr{{Op: OpExit}}, Labels: map[string]int{}, At: map[int]string{}, Tables: map[string]int{}}
}

// Area is area's script, Empty if the game has no such block.
func (l *Library) Area(area int) (*Script, error) {
	names, err := ReadNames(l.Data, area, false, l.Version)
	if err != nil {
		return nil, err
	}
	path := AreaPath(l.Dir, area)
	code := l.block(area)
	if code == nil {
		return Empty(area), nil
	}
	key, err := l.key(code)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	stale := errors.Is(err, fs.ErrNotExist)
	if err != nil && !stale {
		return nil, err
	}
	if !stale {
		_, err := l.Version.checkHeader(string(b))
		stale = errors.Is(err, ErrStale) || textKey(string(b)) != key
	}
	if stale {
		text, _, err := importArea(l.Version, l.Source, l.Data, area, code)
		if err != nil {
			return nil, fmt.Errorf("area %d: %w", area, err)
		}
		text = withKey(text, key)
		if err := writeAtomic(path, text); err != nil {
			return nil, err
		}
		b = []byte(text)
	}
	s, err := Parse(string(b), names)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", path, err)
	}
	s.Area = area
	if err := resolve(s, l.Data, l.Version); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return s, nil
}

// block is area's code, from the first disk that has it.
func (l *Library) block(area int) []byte {
	for disk := 1; disk < maxDisk; disk++ {
		if b := l.Source.Block(disk, area); len(b) >= 2 {
			return b[2:]
		}
	}
	return nil
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
	defer os.Remove(f.Name())
	if _, err := f.WriteString(text); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

// CacheDir is where a game with no scripts folder of its own keeps its
// imported texts: they are made from the game's data, which may be read-only.
func CacheDir(game string) (string, error) {
	d, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "gobox", "scripts", game), nil
}

// Assemble makes a script of raw byte code with no header, one command per
// command, for a test to run a few commands: command numbers are its jumps.
func Assemble(v *Version, code []byte) *Script {
	s := Empty(0)
	s.Code = nil
	r := &reader{code: code, base: v.Base}
	for pc := v.Base; pc < v.Base+len(code); {
		l := v.decode(r, pc)
		s.Code = append(s.Code, Instr{Op: l.Cmd.Op, Args: l.Args})
		pc = l.Next
	}
	return s
}
