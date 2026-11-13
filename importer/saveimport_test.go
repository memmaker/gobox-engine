package importer

import (
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gobox/adventure"
	"gobox/rules1e"
)

// An original lettered save in the saves becomes the slot GAME A, with its
// members' files, holding the original's state; the original is gone.
func TestImportSaves(t *testing.T) {
	scripts, err := filepath.Abs("../data/por/scripts")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(scripts); err != nil {
		t.Skip("no imported POR scripts:", err)
	}
	adv := t.TempDir()
	if err := os.Symlink(scripts, filepath.Join(adv, scriptsDir)); err != nil {
		t.Fatal(err)
	}
	tb, err := json.Marshal(adventure.GameTables{OriginalSave: porSave})
	if err != nil {
		t.Fatal(err)
	}
	saves := filepath.Join(adv, adventure.SavesDir)
	for path, b := range map[string][]byte{filepath.Join(adv, adventure.DataDir, adventure.GameTablesFile): tb} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	f := rules1e.OriginalSave{Set: 1, X: 7, Y: 9}
	f.Area.AreaKind = 20
	f.Game.Shop = 1
	f.Count = 1
	f.Names[0].Set("CHRDATA1.SAV")
	b, err := binary.Append(nil, binary.LittleEndian, &f)
	if err != nil {
		t.Fatal(err)
	}
	b = rules1e.AppendPurse(b, rules1e.Purse{Gold: 12})
	if err := os.MkdirAll(saves, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{"SAVGAMA.DAT": b, "CHRDATA1.SAV": {1}, "CHRDATA1.ITM": {2}, "OTHER.SAV": {3}} {
		if err := os.WriteFile(filepath.Join(saves, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := importSaves(adv); err != nil {
		t.Fatal(err)
	}
	slot := filepath.Join(saves, adventure.SlotsDir, "GAME A")
	g, err := adventure.ReadSave(filepath.Join(slot, adventure.GameFile))
	if err != nil {
		t.Fatal(err)
	}
	var r rules1e.SaveBlob
	if err := json.Unmarshal(g.Rules, &r); err != nil {
		t.Fatal(err)
	}
	if g.X != 7 || g.Y != 9 || r.Area.AreaKind != 20 || r.Game.Shop != 1 || r.Purse.Gold != 12 {
		t.Errorf("imported %+v", g)
	}
	for name, want := range map[string]bool{"CHRDATA1.SAV": true, "CHRDATA1.ITM": true, "OTHER.SAV": false} {
		if _, err := os.Stat(filepath.Join(slot, name)); (err == nil) != want {
			t.Errorf("%s in the slot: %v, want %v", name, err == nil, want)
		}
	}
	if _, err := os.Stat(filepath.Join(saves, "SAVGAMA.DAT")); err == nil {
		t.Error("the original save is still in the folder")
	}
}
