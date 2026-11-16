package importer

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/adventure"
	"gobox/ecl"
	"gobox/engine"
)

// BR2 through ImportBREngine, as BR1 (br1_engine_import_test.go).

func TestImportBREngineFolderBR2(t *testing.T) {
	checkEngineFolder(t, "br2", BR2EXE)
	tb, err := adventure.ReadGameTables(brEngineFolder(t, "br2", BR2EXE))
	if err != nil || tb.Logbook == nil {
		t.Error("br2's logbook:", tb.Logbook, err)
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
