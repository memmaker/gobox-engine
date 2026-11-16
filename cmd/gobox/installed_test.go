package main

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/audio"
	"gobox/importer"
	"gobox/ui"
)

// TestInstalledFoldersLaunch: every game folder under data/ is read and launched as `gobox <id>` does, and its first
// frames come up: each game is a data folder on the shared engine (a stale folder fails here: re-import it).
func TestInstalledFoldersLaunch(t *testing.T) {
	audio.Silent = true
	dirs, _ := filepath.Glob(filepath.Join("..", "..", "data", "*", "manifest.json"))
	if len(dirs) == 0 {
		t.Skip("no installed games")
	}
	for _, m := range dirs {
		dir := filepath.Dir(m)
		t.Run(filepath.Base(dir), func(t *testing.T) {
			if _, err := os.Stat(filepath.Join(dir, "scripts")); err != nil {
				t.Skip("not imported:", err)
			}
			f, err := importer.Read(dir)
			if err != nil {
				t.Fatal(err)
			}
			s := start{Adventure: f.Dir, Data: filepath.Join(f.Dir, "sources")}
			s.Capture.Dev.Saves = t.TempDir()
			l, err := launch(s, f)
			if err != nil {
				t.Fatal(err)
			}
			l.Apply(ui.HostOptions{})
			l.Front.Start("")
			for range 5 {
				l.Front.Frame(1280, 800)
			}
		})
	}
}
