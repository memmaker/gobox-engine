package engine

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/adventure"
	"gobox/importer"
	"gobox/rulesxxvc"
)

// brGame is a Buck Rogers game the shared engine plays: its id, the EXE's offsets and the folder of its originals.
type brGame struct {
	ID   string
	EXE  importer.BREXE
	Name string // for skip messages
}

var (
	gameBR1 = brGame{"br1", importer.BR1EXE, "BR1"}
	gameBR2 = brGame{"br2", importer.BR2EXE, "BR2"}
)

// dos is the originals' folder.
func (g brGame) dos() string { return filepath.Join("..", "data", g.ID, "sources", "dos") }

// brEngine is the shared engine on the game's folder, imported from the originals into a temp dir
// (importer.ImportBR), with the XXVc ruleset, its save folder seeded with the shipped characters as cmd/gobox
// homeSaves does. It skips without the originals.
func brEngine(t *testing.T, g brGame) *Engine {
	t.Helper()
	if _, err := os.Stat(filepath.Join(g.dos(), "START.EXE")); err != nil {
		t.Skip("no "+g.Name+" originals:", err)
	}
	adv := t.TempDir()
	if err := importer.ImportBR(g.ID, g.EXE, g.dos(), adv); err != nil {
		t.Fatal(err)
	}
	return brEngineIn(t, adv)
}

// brEngineIn is a new engine on an imported folder (the import is read, never written, by an engine).
func brEngineIn(t testing.TB, adv string) *Engine {
	t.Helper()
	tables, err := adventure.ReadGameTables(adv)
	if err != nil {
		t.Fatal(err)
	}
	saves := t.TempDir()
	if err := os.CopyFS(saves, os.DirFS(filepath.Join(adv, importer.SavesDir))); err != nil { // a new save folder is seeded from the adventure's saves/ (cmd/gobox homeSaves)
		t.Fatal(err)
	}
	e, err := New(rulesxxvc.Builder, tables, adv, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e.Cfg.SaveDir = saves
	seed := uint32(7)
	e.FixedSeed = &seed
	return e
}

// BR1's names, thin: other tests use them.
func br1Engine(t *testing.T) *Engine               { return brEngine(t, gameBR1) }
func br1EngineIn(t testing.TB, adv string) *Engine { return brEngineIn(t, adv) }

// brBegun is the game begun with one roster character, the intro's pages answered.
func brBegun(t *testing.T, g brGame) *Engine { t.Helper(); return br1BeginOn(t, brEngine(t, g)) }
