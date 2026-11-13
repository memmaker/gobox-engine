package importer

import (
	"gobox/adventure"
	"path/filepath"
	"testing"
)

// TestGameDataFile: sotsb's data goes through data/gamedata.json to the shared
// engine's reader, with its AdLib music (no Amiga files in the folder).
func TestSotsbGameDataFile(t *testing.T) {
	dir := t.TempDir()
	if err := ImportGameTablesTo(filepath.Join(dir, adventure.DataDir)); err != nil {
		t.Fatal(err)
	}
	got, err := adventure.ReadGameData(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "sotsb" || got.AdLib == nil || got.AmigaMusic != nil || len(got.Rests) != 2 || len(got.QuickParty) == 0 {
		t.Errorf("game data %+v", got)
	}
}
