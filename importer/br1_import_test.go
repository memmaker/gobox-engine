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

func br1Folder(t *testing.T) string { return brFolder(t, "br1", BR1EXE) }

// brFolder imports game's originals with ImportBR into a temp folder.
func brFolder(t *testing.T, game string, exe BREXE) string {
	t.Helper()
	_, dos, _ := brBlocks(t, game) // skips without the originals
	dst := t.TempDir()
	if err := ImportBR(game, exe, dos, dst); err != nil {
		t.Fatal(err)
	}
	return dst
}

func TestImportBR1Folder(t *testing.T) {
	checkFolder(t, "br1", BR1EXE)
	d, err := adventure.ReadGameTables(br1Folder(t))
	if err != nil || d.Logbook == nil {
		t.Error("br1's logbook:", d.Logbook, err)
	}
}

// checkFolder is the folder test of both games: no partial layout, none of the files only the old engines read, every area loads with the full one, the start tables.
func checkFolder(t *testing.T, game string, exe BREXE) {
	_, _, blocks := brBlocks(t, game)
	dst := brFolder(t, game, exe)
	scripts := filepath.Join(dst, "scripts")
	if _, err := os.Stat(filepath.Join(scripts, ecl.GlobalLayout)); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(filepath.Join(scripts, ecl.PartialLayout)); !os.IsNotExist(err) {
		t.Errorf("partial layout remains: %v", err)
	}
	for _, gone := range []string{"MON0.json", "ITEM0.json", "SHIPS.bin", "ITEMS.bin", "START.bin", "OVR.bin", "8X8D1.bin"} {
		if _, err := os.Stat(filepath.Join(dst, adventure.DataDir, gone)); !os.IsNotExist(err) {
			t.Errorf("data/%s written: %v", gone, err)
		}
	}
	if m, _ := filepath.Glob(filepath.Join(dst, "saves", "CHRDATA*")); len(m) > 0 {
		t.Errorf("the originals' CHRDATA copies remain: %v", m)
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
	dst := br1Folder(t)
	tb, err := adventure.ReadGameTables(dst)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.New(nil, tb, dst, t.TempDir()); err != nil {
		t.Fatal(err)
	}
}

func TestImportBRUnknownGame(t *testing.T) {
	if err := ImportBR("br3", BREXE{}, t.TempDir(), t.TempDir()); err == nil {
		t.Error("a game without a full layout: want an error")
	}
}

// TestRunBR1: `gobox import br1` (Run) writes the folder ImportBR writes, its manifest the default
// import's, and the front's data in gamedata.json (br1GameData).
func TestRunBR1(t *testing.T) {
	_, dos, _ := br1Blocks(t)
	root := t.TempDir()
	if err := Run("br1", dos, root); err != nil {
		t.Fatal(err)
	}
	f, err := Read(filepath.Join(root, "br1"))
	if err != nil {
		t.Fatal(err)
	}
	if f.Ruleset != "rulesxxvc" || f.Manifest.Start == nil {
		t.Errorf("manifest: ruleset %q, start %v", f.Ruleset, f.Manifest.Start)
	}
	if st := f.Manifest.Start; len(st.Credits) == 0 || len(st.Backdrop) != 1 || st.Backdrop[0].File != "TITLE" || st.Backdrop[0].Block != 4 { // TITLE 4 is the backdrop under the team menu (O030:05BE), not a step
		t.Errorf("title sequence: %+v", st)
	}
	d, err := adventure.ReadGameData(f.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if d.Series != brSeries || d.ID != "br1" || d.Border == nil || len(d.PartyMenu) != len(brPartyMenu) || d.Coin == "" || len(d.DaxMask) == 0 || len(d.QuickParty) != 6 {
		t.Errorf("gamedata.json lacks the front's data: %+v", d)
	}
}
