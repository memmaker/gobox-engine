package importer

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/asset"
	"gobox/asset/orig"
)

// TestImportEqualsDecode imports the originals in ../data/dos and checks the
// folder: MUSIC.DAX's songs are there, its contents check, the art equals
// data/assets/art (the art import).
func TestPodImportEqualsDecode(t *testing.T) {
	dos := filepath.Join("..", "data", "pod", "sources", "dos")
	if _, err := os.Stat(dos); err != nil {
		t.Skip("no originals:", err)
	}
	root := t.TempDir()
	if err := Run("pod", dos, root); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "pod")
	for _, f := range []string{"1.mid", "1.opl.json", "8.mid"} { // MUSIC.DAX's songs (glbr.WriteSongs; TestMIDIEqualsAdLib checks them)
		if _, err := os.Stat(filepath.Join(dst, orig.MusicDir, f)); err != nil {
			t.Error(err)
		}
	}
	if err := CheckContents(dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dst, ManifestFile)); err != nil {
		t.Fatal(err)
	}
	if err := SameTree(filepath.Join(dst, "art"), asset.Dir(dos)); err != nil {
		t.Error("art:", err)
	}
}
