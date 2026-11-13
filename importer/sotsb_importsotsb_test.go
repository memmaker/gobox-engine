package importer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"gobox/asset"
)

// TestImportEqualsDecode imports the originals in ../data/dos and checks the
// folder: the scripts, saves and AdLib music are there, the art equals
// data/assets/art (the art import).
func TestSotsbImportEqualsDecode(t *testing.T) {
	dos := filepath.Join("..", "..", "data", "sotsb", "sources", "dos")
	if _, err := os.Stat(dos); err != nil {
		t.Skip("no originals:", err)
	}
	root := t.TempDir()
	if err := Run("sotsb", dos, root); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "sotsb")
	if err := CheckContents(dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dst, ManifestFile)); err != nil {
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
	if err := SameTree(filepath.Join(dst, "art"), asset.Dir(dos)); err != nil {
		t.Error("art:", err)
	}
	want, err := os.ReadFile("sotsb/sounddata.json")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(filepath.Join(dst, "data", SotsbSoundFile)); err != nil || !bytes.Equal(got, want) {
		t.Error("sound tables differ from sounddata.json", err)
	}
}
