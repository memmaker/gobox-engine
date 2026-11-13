// fonts.go is the shared pool of PC BIOS 8x8 fonts (The Ultimate
// Oldschool PC Font Pack, CC BY-SA 4.0, see LICENSE.txt): code page 437,
// 256 glyphs of 8 rows, bit 7 leftmost, beside the games' own fonts the
// importer pools in the data root (FontPoolDir).

package asset

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
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

// FontBlock is the block of a game's 8X8D*.DAX holding its text font: 64
// upper-case glyphs of 8 rows ('@' first, a character c at c mod 64), then
// the symbol tiles.
const FontBlock = 201

// F8 is a game's text font block in the pool's 256-glyph form: the glyph
// the game draws for ch (upper(ch) mod 64) at every ASCII 0x20..0x7F, so
// a-z hold the capitals (upper case is the font's, not the text's), symbol
// tile k (from 64) at slot 0x80+k-64, the rest blank.
func F8(b []byte) []byte {
	out := make([]byte, 256*8)
	for ch := 0x20; ch < 0x80; ch++ {
		c := ch
		if c >= 'a' && c <= 'z' {
			c -= 0x20
		}
		if i := c % 0x40 * 8; i+8 <= len(b) {
			copy(out[ch*8:], b[i:i+8])
		}
	}
	for k := 64; k*8+8 <= len(b) && 0x80+k-64 < 256; k++ {
		copy(out[(0x80+k-64)*8:], b[k*8:k*8+8])
	}
	return out
}

// FontPoolDir is the font pool's folder in the data root, shared by all
// adventures: each imported game's text font as <Name>.f8, one file per
// distinct font (an embedded font is not repeated there), and
// fontPoolGames naming each game's own.
const FontPoolDir = "fonts"

// fontPoolGames is the pool's file mapping each game ID to its font's name.
const fontPoolGames = "games.json"

// PoolFonts are root's pooled fonts, by name, and the name of game id's
// own font ("": none imported). No pool: none.
func PoolFonts(root, id string) ([]Font, string, error) {
	dir := filepath.Join(root, FontPoolDir)
	ents, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	var out []Font
	for _, e := range ents {
		if filepath.Ext(e.Name()) != ".f8" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, "", err
		}
		if len(b) != 256*8 {
			return nil, "", fmt.Errorf("%s: %d bytes, not 256 glyphs", filepath.Join(dir, e.Name()), len(b))
		}
		out = append(out, Font{strings.TrimSuffix(e.Name(), ".f8"), b})
	}
	games, err := poolGames(dir)
	if err != nil {
		return nil, "", err
	}
	return out, games[id], nil
}

// poolGames is the pool's game-to-font map (none yet: empty).
func poolGames(dir string) (map[string]string, error) {
	games := map[string]string{}
	b, err := os.ReadFile(filepath.Join(dir, fontPoolGames))
	if errors.Is(err, fs.ErrNotExist) {
		return games, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &games); err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Join(dir, fontPoolGames), err)
	}
	return games, nil
}

// AddFont makes bits (F8's form) game id's own font in root's pool: a font
// of the same bytes, pooled or embedded, is reused, else it is written as
// name (a name taken by other bytes is an error).
func AddFont(root, id, name string, bits []byte) error {
	pooled, _, err := PoolFonts(root, id)
	if err != nil {
		return err
	}
	dir := filepath.Join(root, FontPoolDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	all := append(pooled, All()...)
	if i := slices.IndexFunc(all, func(f Font) bool { return bytes.Equal(f.Bits, bits) }); i >= 0 {
		name = all[i].Name
	} else if Find(all, name) >= 0 {
		return fmt.Errorf("font pool %s: %s is another font already", dir, name)
	} else if err := os.WriteFile(filepath.Join(dir, name+".f8"), bits, 0o644); err != nil {
		return err
	}
	games, err := poolGames(dir)
	if err != nil {
		return err
	}
	games[id] = name
	b, err := json.MarshalIndent(games, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, fontPoolGames), b, 0o644)
}
