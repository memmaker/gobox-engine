package importsotsb

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"gobox/asset"
	"gobox/asset/orig"
	"gobox/importer"
)

// TestImportEqualsDecode imports the originals in ../data/dos and checks the
// folder: the scripts, saves and AdLib music are there, the art equals
// data/assets/art (dax.ImportArtIfStale).
func TestImportEqualsDecode(t *testing.T) {
	dos := filepath.Join("..", "..", "..", "data", "sotsb", "sources", "dos")
	if _, err := os.Stat(dos); err != nil {
		t.Skip("no originals:", err)
	}
	root := t.TempDir()
	if err := importer.Run("sotsb", dos, root); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "sotsb")
	if err := importer.CheckContents(dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dst, importer.ManifestFile)); err != nil {
		t.Fatal(err)
	}
	if texts, err := os.ReadDir(filepath.Join(dst, "scripts")); err != nil || len(texts) == 0 {
		t.Fatalf("scripts: %d texts, %v", len(texts), err)
	}
	if _, err := os.Stat(filepath.Join(dst, "saves", "SAVGAMA.DAT")); err != nil {
		t.Error("saves:", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "sound", "SOUND.DAX")); err != nil {
		t.Error("sound:", err)
	}
	if err := orig.ImportArtIfStale(dos, filepath.Join(filepath.Dir(dos), "amiga"), asset.Dir(dos)); err != nil {
		t.Fatal(err)
	}
	if err := importer.SameTree(filepath.Join(dst, "art"), filepath.Join("..", "..", "..", "data", "sotsb", "sources", "assets", "art")); err != nil {
		t.Error("art:", err)
	}
	want, err := os.ReadFile("sounddata.json")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(filepath.Join(dst, "data", SoundFile)); err != nil || !bytes.Equal(got, want) {
		t.Error("sound tables differ from sounddata.json", err)
	}
}
