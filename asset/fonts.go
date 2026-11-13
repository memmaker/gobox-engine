// fonts.go is the shared pool of PC BIOS 8x8 fonts (The Ultimate
// Oldschool PC Font Pack, CC BY-SA 4.0, see LICENSE.txt): code page 437,
// 256 glyphs of 8 rows, bit 7 leftmost. They are not game art; a game picks
// one by name.

package asset

import (
	"embed"
	"sort"
	"strings"
)

//go:embed fonts/*.f8
var fontFS embed.FS

// Font is one 8x8 font: Bits holds 256 glyphs of 8 bytes.
type Font struct {
	Name string
	Bits []byte
}

// All are the embedded fonts, sorted by name ignoring case.
func All() []Font {
	ents, _ := fontFS.ReadDir("fonts")
	var out []Font
	for _, e := range ents {
		b, _ := fontFS.ReadFile("fonts/" + e.Name())
		out = append(out, Font{strings.TrimSuffix(e.Name(), ".f8"), b})
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out
}

// Get is the embedded font called name (no ".f8"), nil if there is none.
func Get(name string) []byte {
	b, _ := fontFS.ReadFile("fonts/" + name + ".f8")
	return b
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
