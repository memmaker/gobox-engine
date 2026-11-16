package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gobox/audio"
	"gobox/importer"
	"gobox/ui"
)

// brLaunch is `gobox import <id> <originals>` into a temp dir and the folder's launch as `gobox <id>` makes it
// (skipped without the originals in data/<id>/sources/dos).
func brLaunch(t *testing.T, id string) (importer.Folder, ui.Launch) {
	t.Helper()
	dos := filepath.Join("..", "..", "data", id, "sources", "dos")
	if _, err := os.Stat(filepath.Join(dos, "START.EXE")); err != nil {
		t.Skip("no "+id+" originals:", err)
	}
	audio.Silent = true
	root := t.TempDir()
	if err := importer.Run(id, dos, root); err != nil {
		t.Fatal(err)
	}
	f, err := importer.Read(filepath.Join(root, id))
	if err != nil {
		t.Fatal(err)
	}
	s := start{Adventure: f.Dir, Data: filepath.Join(f.Dir, "sources")}
	s.Capture.Dev.Saves = t.TempDir()
	l, err := launch(s, f)
	if err != nil {
		t.Fatal(err)
	}
	return f, l
}

// TestLaunchBR1: `gobox import br1` then `gobox br1`: the folder's launch goes through the pool games' code (playPool) on
// the rulesxxvc host, and its Apply takes the options.
func TestLaunchBR1(t *testing.T) { checkLaunch(t, "br1") }

// TestLaunchBR2: the same for Buck Rogers: Matrix Cubed.
func TestLaunchBR2(t *testing.T) { checkLaunch(t, "br2") }

func checkLaunch(t *testing.T, id string) {
	f, l := brLaunch(t, id)
	if l.Front == nil || l.Title != f.Title || l.Apply == nil {
		t.Fatalf("launch %+v", l)
	}
	l.Apply(ui.HostOptions{})
}

// TestStartupTimingBR1: start-up as main does it, headless: the launch and the first frames are quick.
func TestStartupTimingBR1(t *testing.T) {
	_, l := brLaunch(t, "br1")
	start := time.Now()
	l.Front.Start("")
	for range 5 {
		l.Front.Frame(1280, 800)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("the first five frames took %v, want well under a second", d)
	}
}
