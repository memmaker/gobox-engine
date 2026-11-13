package script

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gobox/dax"
)

// headerWord starts a script text's first line: "version <id> <revision>
// [<key>]"; a Library's text carries its key (Library.key).
const headerWord = "version"

// Header is the first line of the text this version writes.
func (v *Version) Header() string {
	return fmt.Sprintf("%s %d %d\n", headerWord, v.ID, FormatRevision)
}

// ErrStale: the text was made by another version or format revision, so the
// import must run again.
var ErrStale = errors.New("script text from another decompiler version")

// checkHeader splits a text into its header and the rest; the rest keeps an
// empty first line so that line numbers still count the file's.
func (v *Version) checkHeader(text string) (string, error) {
	first, rest, _ := strings.Cut(text, "\n")
	f := strings.Fields(first)
	if (len(f) != 3 && len(f) != 4) || f[0] != headerWord {
		return "", fmt.Errorf("%w: no version line", ErrStale)
	}
	id, err1 := strconv.Atoi(f[1])
	rev, err2 := strconv.Atoi(f[2])
	if err1 != nil || err2 != nil || !v.Current(VersionID(id), rev) {
		return "", fmt.Errorf("%w: %q", ErrStale, first)
	}
	return "\n" + rest, nil
}

// textKey is the key on text's first line ("" for none).
func textKey(text string) string {
	first, _, _ := strings.Cut(text, "\n")
	if f := strings.Fields(first); len(f) == 4 && f[0] == headerWord {
		return f[3]
	}
	return ""
}

// withKey is text with key on its first line.
func withKey(text, key string) string {
	first, rest, _ := strings.Cut(text, "\n")
	f := strings.Fields(first)
	return strings.Join(append(f[:min(len(f), 3)], key), " ") + "\n" + rest
}

// Source is the game data an import reads.
type Source struct {
	Block    func(disk, block int) []byte // an ECL block with its 2-byte header; nil when absent
	Treasure func(set, block int) string  // an ITEM block's items, for comments
}

// AreaPath is area's script file in the scripts folder.
func AreaPath(dir string, area int) string {
	return filepath.Join(dir, fmt.Sprintf("area%d.txt", area))
}

const maxDisk, maxBlock = 9, 256

// A Fix is a game's special case in its decompiler: in area's text, Old
// (whole lines, exactly once) becomes New. A fix whose Old is not there
// once, or whose area has no block, fails the import.
type Fix struct {
	Area     int
	Old, New string
}

// Import decompiles every block of the source into dir, with the game's
// script data (names/, layout/) and its special cases. A block on several
// disks is taken from the first. Each text carries Key. It returns the
// folded gates.
func Import(v *Version, src Source, data fs.FS, dir string, fold Folding, fixes []Fix) ([]FoldedGate, error) {
	key, err := Key(v, data, fold, fixes)
	if err != nil {
		return nil, err
	}
	var all []FoldedGate
	done := map[int]bool{}
	for disk := 1; disk < maxDisk; disk++ {
		for block := range maxBlock {
			b := src.Block(disk, block)
			if len(b) < 2 || done[block] {
				continue
			}
			done[block] = true
			text, folds, err := importArea(v, src, data, block, b[2:], fold, fixes)
			if err != nil {
				return nil, fmt.Errorf("area %d: %w", block, err)
			}
			all = append(all, folds...)
			if err := writeAtomic(AreaPath(dir, block), withKey(text, key)); err != nil {
				return nil, err
			}
		}
	}
	for _, f := range fixes {
		if !done[f.Area] {
			return nil, fmt.Errorf("a fix for area %d, which has no block", f.Area)
		}
	}
	return all, nil
}

func importArea(v *Version, src Source, data fs.FS, area int, code []byte, fold Folding, fixes []Fix) (string, []FoldedGate, error) {
	names, err := ReadNames(data, area, true, v)
	if err != nil {
		return "", nil, err
	}
	if src.Treasure != nil {
		names.Treasure = func(block int) string { return src.Treasure(names.Set, block) }
	}
	lay, err := ReadLayout(data, area)
	if err != nil {
		return "", nil, err
	}
	text, folds, err := DecodeBlock(code, area, names, lay, fold)
	if err != nil {
		return "", nil, err
	}
	for i, f := range fixes {
		if f.Area != area {
			continue
		}
		if n := strings.Count("\n"+text, "\n"+f.Old); n != 1 || f.Old == "" || !strings.HasSuffix(f.Old, "\n") {
			return "", nil, fmt.Errorf("fix %d: its site is there %d times, want once", i, n)
		}
		at := strings.Index("\n"+text, "\n"+f.Old)
		text = text[:at] + f.New + text[at+len(f.Old):]
	}
	return text, folds, nil
}

// Stale reports whether dir's scripts need the import: none at all, or
// any made by another decompiler or from other data (Key).
func Stale(v *Version, dir string, data fs.FS, fold Folding, fixes []Fix) (bool, error) {
	key, err := Key(v, data, fold, fixes)
	if err != nil {
		return false, err
	}
	files, err := filepath.Glob(filepath.Join(dir, "area*.txt"))
	if err != nil {
		return false, err
	}
	if len(files) == 0 {
		return true, nil
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return false, err
		}
		if _, err := v.checkHeader(string(b)); errors.Is(err, ErrStale) || textKey(string(b)) != key {
			return true, nil
		}
	}
	return false, nil
}

// ImportFolder is Import into an adventure folder's scripts dir (empty or
// new): the game's script data is copied there first, so the folder reads
// without the game's embedded files.
func ImportFolder(v *Version, src Source, data fs.FS, dir string, fold Folding, fixes []Fix) error {
	if err := os.CopyFS(dir, data); err != nil {
		return err
	}
	_, err := Import(v, src, os.DirFS(dir), dir, fold, fixes)
	return err
}

// ImportIfStale runs Import when dir's scripts are missing or stale (Stale):
// a game's setup step, and a test binary's before it loads an area.
func ImportIfStale(v *Version, src Source, data fs.FS, dir string, fold Folding, fixes []Fix) error {
	stale, err := Stale(v, dir, data, fold, fixes)
	if err != nil || !stale {
		return err
	}
	_, err = Import(v, src, data, dir, fold, fixes)
	return err
}

// DAXSource reads the blocks of the game's ECL*.DAX, for the import.
func DAXSource(l *dax.Loader) Source {
	return Source{Block: func(disk, block int) []byte {
		name := fmt.Sprintf("ECL%d", disk)
		if !l.Exists(name + ".DAX") {
			return nil
		}
		return l.BlockShared(name, block)
	}}
}

// ImportDAXFolder decompiles the ECL*.DAX in dos into an adventure folder's
// scripts dir (ImportFolder with the DAX source and gate folding): the
// games' importer step.
func ImportDAXFolder(v *Version, dos string, data fs.FS, dir string, fixes []Fix) error {
	l, err := dax.NewLoader(dos)
	if err != nil {
		return err
	}
	return ImportFolder(v, DAXSource(l), data, dir, FoldGates, fixes)
}
