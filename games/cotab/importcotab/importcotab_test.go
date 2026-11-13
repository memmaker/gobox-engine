package importcotab

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
// folder: its contents, the art equal to data/assets/art
// (dax.ImportArtIfStale), the sound tables to testdata/s.
func TestImportEqualsDecode(t *testing.T) {
	dos := filepath.Join("..", "..", "..", "data", "cotab", "sources", "dos")
	if _, err := os.Stat(dos); err != nil {
		t.Skip("no originals:", err)
	}
	root := t.TempDir()
	if err := importer.Run("cotab", dos, root); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "cotab")
	if err := importer.CheckContents(dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dst, importer.ManifestFile)); err != nil {
		t.Fatal(err)
	}
	if err := orig.ImportArtIfStale(dos, filepath.Join(filepath.Dir(dos), "amiga"), asset.Dir(dos)); err != nil {
		t.Fatal(err)
	}
	if err := importer.SameTree(filepath.Join(dst, "art"), filepath.Join("..", "..", "..", "data", "cotab", "sources", "assets", "art")); err != nil {
		t.Error("art:", err)
	}
	want, err := os.ReadFile(filepath.Join("testdata", "sounddata.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(filepath.Join(dst, "data", SoundFile)); err != nil || !bytes.Equal(got, want) {
		t.Error("sound tables differ from testdata/sounddata.json", err)
	}
}
