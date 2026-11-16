package importer

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gobox/adventure"
	"gobox/ecl"
	"gobox/engine"
	"gobox/rules"
)

func br1EngineFolder(t *testing.T) string { return brEngineFolder(t, "br1", BR1EXE) }

// brEngineFolder imports game's originals with ImportBREngine into a temp folder.
func brEngineFolder(t *testing.T, game string, exe BREXE) string {
	t.Helper()
	_, dos, _ := brBlocks(t, game) // skips without the originals
	dst := t.TempDir()
	if err := ImportBREngine(game, exe, dos, dst); err != nil {
		t.Fatal(err)
	}
	return dst
}

func TestImportBREngineFolder(t *testing.T) {
	checkEngineFolder(t, "br1", BR1EXE)
	d, err := adventure.ReadGameTables(br1EngineFolder(t))
	if err != nil || d.Logbook == nil {
		t.Error("br1's logbook:", d.Logbook, err)
	}
}

// checkEngineFolder is the folder test of both games: no partial layout, every area loads with the full one, the start tables.
func checkEngineFolder(t *testing.T, game string, exe BREXE) {
	_, _, blocks := brBlocks(t, game)
	dst := brEngineFolder(t, game, exe)
	scripts := filepath.Join(dst, "scripts")
	if _, err := os.Stat(filepath.Join(scripts, ecl.GlobalLayout)); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(filepath.Join(scripts, ecl.PartialLayout)); !os.IsNotExist(err) {
		t.Errorf("partial layout remains: %v", err)
	}
	lib := &ecl.Library{Version: ecl.V4, Dir: scripts, Data: os.DirFS(scripts), Fold: ecl.FoldGates}
	for _, area := range slices.Sorted(maps.Keys(blocks)) {
		if _, err := lib.Area(area); err != nil {
			t.Errorf("area %d: %v", area, err)
		}
	}
	tb, err := adventure.ReadGameTables(dst)
	if err != nil {
		t.Fatal(err)
	}
	if tb.Dialect != rules.DialectBR || tb.Script != ecl.V4 || tb.FirstBlock != 16 || tb.GameSet != 1 ||
		tb.BigPictures.First != 112 || tb.OriginalSave.CodeBase != 32768 || tb.OriginalSave.CodeSize != 9217 ||
		tb.Art == nil || tb.Art.CombatSprites != "COMSPR" || len(tb.Symbols) != 2 {
		t.Errorf("game tables: %+v", tb)
	}
	d, err := adventure.ReadGameData(dst)
	if err != nil || d.ID != game {
		t.Errorf("game data: %q, %v", d.ID, err)
	}
}

// TestBR1EngineNew: engine.New on the folder returns no error. The ruleset is built at its first use, so New needs none
// (nil builder); Init is not called: its first act, resetAdventure's rules.NewAdventure, asks the ruleset.
func TestBR1EngineNew(t *testing.T) {
	dst := br1EngineFolder(t)
	tb, err := adventure.ReadGameTables(dst)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.New(nil, tb, dst, t.TempDir()); err != nil {
		t.Fatal(err)
	}
}

func TestImportBRUnchanged(t *testing.T) {
	_, dos, _ := br1Blocks(t)
	dst := t.TempDir()
	if err := ImportBR("br1", BR1EXE, dos, dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dst, "scripts", ecl.PartialLayout)); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(filepath.Join(dst, adventure.DataDir, adventure.GameTablesFile)); !os.IsNotExist(err) {
		t.Errorf("ImportBR wrote gametables.json: %v", err)
	}
}

func TestImportBREngineUnknownGame(t *testing.T) {
	if err := ImportBREngine("br3", BREXE{}, t.TempDir(), t.TempDir()); err == nil {
		t.Error("a game without a full layout: want an error")
	}
}
