package asset

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFontPool: AddFont pools each distinct font once (an embedded one not
// at all) and names every game's own; PoolFonts reads them back.
func TestFontPool(t *testing.T) {
	root := t.TempDir()
	own := make([]byte, 256*8)
	own['A'*8] = 0x81
	if err := AddFont(root, "por", "Pool_of_Radiance", Get(GameFont)); err != nil {
		t.Fatal(err)
	}
	if err := AddFont(root, "br1", "Buck_Rogers", own); err != nil {
		t.Fatal(err)
	}
	if err := AddFont(root, "br2", "Matrix_Cubed", own); err != nil {
		t.Fatal(err)
	}
	if err := AddFont(root, "x", "Buck_Rogers", make([]byte, 256*8)); err == nil {
		t.Fatal("a taken name of other bytes was pooled")
	}
	for id, want := range map[string]string{"por": GameFont, "br1": "Buck_Rogers", "br2": "Buck_Rogers", "x": ""} {
		l, name, err := PoolFonts(root, id)
		if err != nil || name != want || len(l) != 1 || l[0].Name != "Buck_Rogers" {
			t.Fatalf("%s: %d fonts, own %q, %v", id, len(l), name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, FontPoolDir, GameFont+".f8")); err == nil {
		t.Fatal("an embedded font was pooled")
	}
}
