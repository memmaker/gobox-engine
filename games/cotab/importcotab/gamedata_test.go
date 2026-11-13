package importcotab

import (
	"path/filepath"
	"reflect"
	"testing"

	"gobox/engine"
)

// TestGameDataFile: cotab's StaticData goes through data/gamedata.json to the
// shared engine's reader unchanged (the folder's Amiga music is pool's to keep or drop).
func TestGameDataFile(t *testing.T) {
	dir := t.TempDir()
	want := StaticData()
	if err := engine.WriteGameData(filepath.Join(dir, engine.DataDir), want); err != nil {
		t.Fatal(err)
	}
	got, err := engine.ReadGameData(dir)
	if err != nil {
		t.Fatal(err)
	}
	want.AmigaMusic = nil // no sound/ in dir
	if got.ID != "cotab" || !reflect.DeepEqual(got, want) {
		t.Errorf("read back %+v, want %+v", got, want)
	}
	if got.Border == nil || got.Border.Edges["top"] == nil || got.DaxMask["8X8D1"] != 0xD || len(got.PartyMenu) != 13 {
		t.Errorf("border or masks lost: %+v", got)
	}
}
