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
