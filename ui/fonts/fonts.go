// Package fonts is the shared pool of PC BIOS 8x8 fonts (The Ultimate
// Oldschool PC Font Pack, CC BY-SA 4.0, see LICENSE.txt): code page 437,
// 256 glyphs of 8 rows, bit 7 leftmost. They are not game art; a game picks
// one by name.
package fonts

import (
	"embed"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed *.f8
var fontFS embed.FS

// Font is one 8x8 font: Bits holds 256 glyphs of 8 bytes.
type Font struct {
	Name string
	Bits []byte
}

// All are the embedded fonts, sorted by name ignoring case.
func All() []Font {
	ents, _ := fontFS.ReadDir(".")
	var out []Font
	for _, e := range ents {
		b, _ := fontFS.ReadFile(e.Name())
		out = append(out, Font{strings.TrimSuffix(e.Name(), ".f8"), b})
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out
}

// Get is the embedded font called name (no ".f8"), nil if there is none.
func Get(name string) []byte {
	b, _ := fontFS.ReadFile(name + ".f8")
	return b
}

// Load are the game's own .f8 fonts in dir (2048 bytes each), then the
// embedded ones. A missing dir gives just the embedded fonts.
func Load(dir string) []Font { return append(FontDir(dir), All()...) }

// FontDir are the .f8 fonts (2048 bytes each) in dir.
func FontDir(dir string) []Font {
	var out []Font
	ents, _ := os.ReadDir(dir)
	for _, d := range ents {
		if !strings.EqualFold(filepath.Ext(d.Name()), ".f8") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, d.Name()))
		if err == nil && len(b) == 256*8 {
			out = append(out, Font{strings.TrimSuffix(d.Name(), filepath.Ext(d.Name())), b})
		}
	}
	return out
}

// Find is the index of the font called name in fs, or -1.
func Find(fs []Font, name string) int {
	for i, f := range fs {
		if f.Name == name {
			return i
		}
	}
	return -1
}
