package importer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesFolderAndManifest(t *testing.T) {
	src, root := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(src, "START.EXE"), []byte("exe"), 0o644)
	os.WriteFile(filepath.Join(src, "a.dax"), []byte("dax"), 0o644)
	os.WriteFile(filepath.Join(src, "README"), []byte("no"), 0o644)
	Register(Game{ID: "test", Title: "Test", Ruleset: "rules1e", Sources: []string{"*.DAX", "START.EXE"},
		Import: func(src, dst string) error { return os.WriteFile(filepath.Join(dst, "x.txt"), nil, 0o644) }})
	if err := Main([]string{"-out", root, "test", src}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "test", ManifestFile))
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m.ID != "test" || m.Ruleset != "rules1e" || m.Format != Format || len(m.Source) != 2 || m.Source["START.EXE"] == "" {
		t.Fatalf("manifest %+v", m)
	}
	if _, err := os.Stat(filepath.Join(root, "test", "x.txt")); err != nil {
		t.Fatal(err)
	}
	if err := Run("none", src, root); err == nil {
		t.Fatal("unknown game imported")
	}
}

// TestRunKeepsSourcesAndDocs: an import into a game's data folder replaces
// the import's entries and keeps Kept, its originals among them.
func TestRunKeepsSourcesAndDocs(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "keep")
	src := filepath.Join(dir, "sources", "dos")
	os.MkdirAll(src, 0o755)
	os.MkdirAll(filepath.Join(dir, "docs"), 0o755)
	os.WriteFile(filepath.Join(src, "START.EXE"), []byte("exe"), 0o644)
	os.WriteFile(filepath.Join(dir, "stale.txt"), nil, 0o644)
	Register(Game{ID: "keep", Ruleset: "rules1e", Sources: []string{"START.EXE"},
		Import: func(src, dst string) error { return os.WriteFile(filepath.Join(dst, "x.txt"), nil, 0o644) }})
	if err := Run("keep", src, root); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"sources/dos/START.EXE", "docs", "x.txt", ManifestFile} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Error(err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "stale.txt")); err == nil {
		t.Error("stale.txt kept")
	}
}
