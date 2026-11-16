package importer

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/adventure"
	"gobox/ecl"
	"gobox/engine"
	"gobox/rules"
)

// BR2 through ImportBREngine, as BR1 (br1_engine_import_test.go).

func TestImportBREngineFolderBR2(t *testing.T) {
	checkEngineFolder(t, "br2", BR2EXE)
	tb, err := adventure.ReadGameTables(brEngineFolder(t, "br2", BR2EXE))
	if err != nil || tb.Logbook == nil {
		t.Error("br2's logbook:", tb.Logbook, err)
	}
	if tb.OverlandSteps != (rules.Size{W: 38, H: 15}) || len(tb.Places) == 0 || tb.Walls == nil || len(tb.Walls.Backs) == 0 || tb.Space == nil {
		t.Errorf("br2 game tables lack data: overland %v places %d walls %v space %v", tb.OverlandSteps, len(tb.Places), tb.Walls != nil, tb.Space != nil)
	}
	if len(tb.PictureSets) != 6 || tb.Art.Title != "TITLE" {
		t.Errorf("br2 game tables: %+v", tb)
	}
}

// TestBR2EngineNew: engine.New on the folder returns no error (the ruleset is built at its first use; Init is not called).
func TestBR2EngineNew(t *testing.T) {
	dst := brEngineFolder(t, "br2", BR2EXE)
	tb, err := adventure.ReadGameTables(dst)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.New(nil, tb, dst, t.TempDir()); err != nil {
		t.Fatal(err)
	}
}

func TestImportBRUnchangedBR2(t *testing.T) {
	_, dos, _ := brBlocks(t, "br2")
	dst := t.TempDir()
	if err := ImportBR("br2", BR2EXE, dos, dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dst, "scripts", ecl.PartialLayout)); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(filepath.Join(dst, adventure.DataDir, adventure.GameTablesFile)); !os.IsNotExist(err) {
		t.Errorf("ImportBR wrote gametables.json: %v", err)
	}
}

// TestBR2GameData: the front's data of BR2 is the old BR2 engine's (buckrogers2-go rules/data.go): its party menu (Quick Start before
// Begin Adventuring), masks, the Aid key and its own music places.
func TestBR2GameData(t *testing.T) {
	d, err := adventure.ReadGameData(brEngineFolder(t, "br2", BR2EXE))
	if err != nil {
		t.Fatal(err)
	}
	if d.ID != "br2" || d.Border == nil || d.Coin == "" || len(d.QuickParty) != 6 || d.DaxMask["CURSOR"] != 13 || d.VerbKeys["Aid"] != 'H' {
		t.Errorf("game data: %+v", d)
	}
	if n := len(d.PartyMenu); n != 13 || d.PartyMenu[10].Label != "Quick Start" || d.PartyMenu[11].Label != "Begin Adventuring" {
		t.Errorf("party menu %v", d.PartyMenu)
	}
	if len(d.MusicPlaces) != len(br2MusicPlaces) {
		t.Errorf("music places %v", d.MusicPlaces)
	}
}
