package dax

import (
	"os"
	"testing"
)

// PoD's border pictures (BORDERS.DAX block 0) decode as VGA, masked colour 13.
func TestParseVGAPod(t *testing.T) {
	const dir = "../data/pod/extras/gog/"
	if _, err := os.Stat(dir + "BORDERS.DAX"); err != nil {
		t.Skip("no PoD data")
	}
	f, err := Open(dir + "BORDERS.DAX")
	if err != nil {
		t.Fatal(err)
	}
	im := ParseVGA(f.Block(0), true, 13)
	if im == nil || im.Count == 0 || len(im.Palette) != 256 || im.Palette[13].A != 0 {
		t.Fatalf("%+v", im)
	}
	for _, p := range im.Data {
		if im.Palette[p].A == 0 && p != 13 {
			t.Fatal("stray transparent")
		}
	}
}
