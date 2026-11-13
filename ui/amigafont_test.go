package ui

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/asset"
	"gobox/game"
)

// TestAmigaFont: the game's own font follows the art switch (Front.Arts):
// the pool's Amiga twin's font while the Amiga art shows, else the DOS one.
func TestAmigaFont(t *testing.T) {
	defer func(a *asset.Art, am bool) { daxArt.art, daxArt.amiga = a, am; clear(daxPics) }(daxArt.art, daxArt.amiga)
	root, art := t.TempDir(), t.TempDir()
	if err := os.Mkdir(filepath.Join(art, string(asset.Amiga)), 0o755); err != nil {
		t.Fatal(err)
	}
	daxArt.art, daxArt.amiga = asset.OpenArt(art), false
	font := func(row byte) []byte {
		b := make([]byte, 256*8)
		b['A'*8] = row
		return b
	}
	if err := asset.AddFont(root, "g", "G_DOS", font(0x81)); err != nil {
		t.Fatal(err)
	}
	if err := asset.AddFont(root, asset.PlatformID("g", asset.PlatformAmiga), "G_Amiga", font(0x42)); err != nil {
		t.Fatal(err)
	}
	f := &Front{}
	f.chooseFonts(game.Data{Adventure: filepath.Join(root, "g")})
	_, set := f.Arts()
	if set == nil {
		t.Fatal("no art switch")
	}
	for _, c := range []struct {
		dos bool
		row byte
	}{{true, 0x81}, {false, 0x42}, {true, 0x81}} {
		set(c.dos)
		if g := f.Glyph('A'); g[0] != c.row {
			t.Errorf("DOS art %v: 'A' row 0 %#x, want %#x", c.dos, g[0], c.row)
		}
	}
}
