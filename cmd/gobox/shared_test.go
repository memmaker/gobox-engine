package main

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/audio"
	"gobox/importer"
	"gobox/ui"
)

// TestSharedLaunchBR1: `gobox import -shared br1` then `gobox -shared br1`: the folder's launch goes through the
// pool games' code (playPool) on the rulesxxvc host, and its Apply takes the options. Without -shared the same
// folder goes to the old engine (playOther), as the pool games go to playPool either way.
func TestSharedLaunchBR1(t *testing.T) { sharedLaunch(t, "br1") }

// TestSharedLaunchBR2: the same for Buck Rogers: Matrix Cubed.
func TestSharedLaunchBR2(t *testing.T) { sharedLaunch(t, "br2") }

func sharedLaunch(t *testing.T, id string) {
	dos := filepath.Join("..", "..", "data", id, "sources", "dos")
	if _, err := os.Stat(filepath.Join(dos, "START.EXE")); err != nil {
		t.Skip("no "+id+" originals:", err)
	}
	audio.Silent = true
	root := t.TempDir()
	if err := importer.RunWith(id, dos, root, true); err != nil {
		t.Fatal(err)
	}
	f, err := importer.Read(filepath.Join(root, id))
	if err != nil {
		t.Fatal(err)
	}
	s := start{Adventure: f.Dir, Data: filepath.Join(f.Dir, "sources"), Shared: true}
	s.Capture.Dev.Saves = t.TempDir()
	l, err := launch(s, f)
	if err != nil {
		t.Fatal(err)
	}
	if l.Front == nil || l.Title != f.Title || l.Apply == nil {
		t.Fatalf("launch %+v", l)
	}
	l.Apply(ui.HostOptions{})
}
