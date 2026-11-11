package importer

import (
	"os"
	"path/filepath"
	"testing"
)

// TestList: Run's folder is listed with its manifest; a folder with another
// format, a foreign ID or no manifest is not.
func TestList(t *testing.T) {
	src, root := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "GAME.DAT"), []byte{1}, 0o644); err != nil {
		t.Fatal(err)
	}
	games["listed"] = Game{ID: "listed", Title: "Listed", Ruleset: "rules1e", Sources: []string{"*.DAT"}, Import: func(_, _ string) error { return nil }}
	defer delete(games, "listed")
	if err := Run("listed", src, root); err != nil {
		t.Fatal(err)
	}
	bad := map[string]string{
		"old":     `{"ID":"old","Ruleset":"rules1e","Format":1}`,
		"foreign": `{"ID":"other","Ruleset":"rules1e","Format":2}`,
	}
	for dir, m := range bad {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, dir, ManifestFile), []byte(m), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "none"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, skipped, err := List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "listed" || got[0].Title != "Listed" || got[0].Dir != filepath.Join(root, "listed") {
		t.Errorf("listed %+v", got)
	}
	if len(skipped) != 3 {
		t.Errorf("skipped %v", skipped)
	}
}
