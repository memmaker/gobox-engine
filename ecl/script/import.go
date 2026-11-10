package script

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// patchPath is area's import patch: the hand fixes, a unified diff against
// the decompiler's output.
func patchPath(area int) string {
	return fmt.Sprintf("patches/area%d.patch", area)
}

const maxDisk, maxBlock = 9, 256

// Import decompiles every block of the source into dir (with its names/ and
// patches/ folders), applying each area's patch. A block on several disks is
// taken from the first. It returns the folded gates.
func Import(v *Version, src Source, dir string) ([]FoldedGate, error) {
	var all []FoldedGate
	done := map[int]bool{}
	for disk := 1; disk < maxDisk; disk++ {
		for block := range maxBlock {
			b := src.Block(disk, block)
			if len(b) < 2 || done[block] {
				continue
			}
			done[block] = true
			text, folds, err := importArea(v, src, os.DirFS(dir), block, b[2:])
			if err != nil {
				return nil, fmt.Errorf("area %d: %w", block, err)
			}
			all = append(all, folds...)
			if err := os.WriteFile(AreaPath(dir, block), []byte(text), 0o644); err != nil {
				return nil, err
			}
		}
	}
	return all, nil
}

func importArea(v *Version, src Source, data fs.FS, area int, code []byte) (string, []FoldedGate, error) {
	names, err := ReadNames(data, area, true, v)
	if err != nil {
		return "", nil, err
	}
	if src.Treasure != nil {
		names.Treasure = func(block int) string { return src.Treasure(names.Set, block) }
	}
	text, folds, err := DecodeBlock(code, area, names)
	if err != nil {
		return "", nil, err
	}
	p, err := fs.ReadFile(data, patchPath(area))
	if errors.Is(err, fs.ErrNotExist) {
		return text, folds, nil
	} else if err != nil {
		return "", nil, err
	}
	text, err = ApplyPatch(text, string(p))
	return text, folds, err
}

// Stale reports whether dir's scripts need the import: any existing area
// script (or none at all) of another version.
func Stale(v *Version, dir string) (bool, error) {
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
		if _, err := v.checkHeader(string(b)); errors.Is(err, ErrStale) {
			return true, nil
		}
	}
	return false, nil
}

// ApplyPatch applies a unified diff to text, line for line: any context or
// removed line that does not match the text is an error.
func ApplyPatch(text, patch string) (string, error) {
	src := strings.SplitAfter(text, "\n")
	var out []string
	at := 0 // lines of src consumed
	lines := strings.SplitAfter(patch, "\n")
	for i := 0; i < len(lines); {
		l := lines[i]
		if !strings.HasPrefix(l, "@@ ") {
			i++ // a header line
			continue
		}
		var from, nFrom, to, nTo int
		if _, err := fmt.Sscanf(l, "@@ -%d,%d +%d,%d @@", &from, &nFrom, &to, &nTo); err != nil {
			return "", fmt.Errorf("patch line %d: hunk header %q", i+1, l)
		}
		if nFrom == 0 {
			from++ // diff's "after line N"
		}
		if from-1 < at || from-1 > len(src) {
			return "", fmt.Errorf("patch line %d: hunk out of order", i+1)
		}
		out = append(out, src[at:from-1]...)
		at = from - 1
		for i++; i < len(lines) && !strings.HasPrefix(lines[i], "@@ ") && lines[i] != ""; i++ {
			body := lines[i]
			switch body[0] {
			case ' ', '-':
				if at >= len(src) || src[at] != body[1:] {
					return "", fmt.Errorf("patch line %d does not match line %d of the decompiled text", i+1, at+1)
				}
				if body[0] == ' ' {
					out = append(out, body[1:])
				}
				at++
			case '+':
				out = append(out, body[1:])
			case '\\': // "\ No newline at end of file"
			default:
				return "", fmt.Errorf("patch line %d: %q", i+1, body)
			}
		}
	}
	return strings.Join(append(out, src[at:]...), ""), nil
}
