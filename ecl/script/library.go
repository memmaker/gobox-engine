package script

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gobox/dax"
)

// Library is a game's decompiled scripts: one text per area in Dir, each
// made from the game's data the first time it is needed or when the
// decompiler has moved on.
type Library struct {
	Version *Version
	Source  Source
	Dir     string
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
	names, err := ReadNames(l.Dir, area, false, l.Version)
	if err != nil {
		return nil, err
	}
	path := AreaPath(l.Dir, area)
	b, err := os.ReadFile(path)
	stale := errors.Is(err, fs.ErrNotExist)
	if err != nil && !stale {
		return nil, err
	}
	if !stale {
		_, err := l.Version.checkHeader(string(b))
		stale = errors.Is(err, ErrStale)
	}
	if stale {
		code := l.block(area)
		if code == nil {
			return Empty(area), nil
		}
		text, _, err := importArea(l.Version, l.Source, l.Dir, area, code)
		if err != nil {
			return nil, fmt.Errorf("area %d: %w", area, err)
		}
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
