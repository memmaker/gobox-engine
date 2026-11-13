package games

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/asset"
	"gobox/asset/orig"
	"gobox/importer"
)

// TestImportEqualsDecode imports the originals in ../data/dos and checks the
// folder: MUSIC.DAX's songs are there, its contents check, the art equals
// data/assets/art (dax.ImportArtIfStale).
func TestPodImportEqualsDecode(t *testing.T) {
	dos := filepath.Join("..", "..", "data", "pod", "sources", "dos")
	if _, err := os.Stat(dos); err != nil {
		t.Skip("no originals:", err)
	}
	root := t.TempDir()
	if err := importer.Run("pod", dos, root); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "pod")
	for _, f := range []string{"1.mid", "1.opl.json", "8.mid"} { // MUSIC.DAX's songs (glbr.WriteSongs; TestMIDIEqualsAdLib checks them)
		if _, err := os.Stat(filepath.Join(dst, orig.MusicDir, f)); err != nil {
			t.Error(err)
		}
	}
	if err := importer.CheckContents(dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dst, importer.ManifestFile)); err != nil {
		t.Fatal(err)
	}
	if err := orig.ImportArtIfStale(dos, filepath.Join(filepath.Dir(dos), "amiga"), asset.Dir(dos)); err != nil {
		t.Fatal(err)
	}
	if err := importer.SameTree(filepath.Join(dst, "art"), filepath.Join("..", "..", "data", "pod", "sources", "assets", "art")); err != nil {
		t.Error("art:", err)
	}
}
