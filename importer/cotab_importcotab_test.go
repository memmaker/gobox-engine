package importer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"gobox/asset"
)

// TestImportEqualsDecode imports the originals in ../data/dos and checks the
// folder: its contents, the art equal to data/assets/art
// (the art import), the sound tables to testdata/s.
func TestCotabImportEqualsDecode(t *testing.T) {
	dos := filepath.Join("..", "..", "data", "cotab", "sources", "dos")
	if _, err := os.Stat(dos); err != nil {
		t.Skip("no originals:", err)
	}
	root := t.TempDir()
	if err := Run("cotab", dos, root); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "cotab")
	if err := CheckContents(dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dst, ManifestFile)); err != nil {
		t.Fatal(err)
	}
	if err := SameTree(filepath.Join(dst, "art"), asset.Dir(dos)); err != nil {
		t.Error("art:", err)
	}
	want, err := os.ReadFile(filepath.Join("cotab", "testdata", "sounddata.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(filepath.Join(dst, "data", CotabSoundFile)); err != nil || !bytes.Equal(got, want) {
		t.Error("sound tables differ from cotab/testdata/sounddata.json", err)
	}
}
