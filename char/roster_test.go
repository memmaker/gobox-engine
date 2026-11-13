package char

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gobox/rules"
	"gobox/rules1e"
)

// TestReadRoster: a later game's characters in its folder come out as canonical files, by name,
// with the record's own flag byte; other files in the folder are not characters.
func TestReadRoster(t *testing.T) {
	dir := t.TempDir()
	var r Record422
	r.PName = [16]byte{4, 'I', 'l', 'a', 'n'}
	r.Pad402[0] = 1
	if err := os.WriteFile(filepath.Join(dir, "ILAN.GUY"), Encode(r), 0o644); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "SAVE.SAV"), Encode(r), 0o644)
	got, err := ReadRoster(rules.Layout422, dir)
	if err != nil || len(got) != 1 {
		t.Fatalf("%d characters, %v", len(got), err)
	}
	var c rules1e.CharacterJSON
	if err := json.Unmarshal(got[0], &c); err != nil {
		t.Fatal(err)
	}
	if c.Sheet.CharName != "Ilan" || c.Sheet.ScriptFlag != 1 || c.Kept == nil || c.Kept.Layout != rules.Layout422 {
		t.Errorf("character %+v", c.Sheet)
	}
	if _, err := ReadRoster(rules.RecordLayout(7), dir); err == nil {
		t.Error("an unknown layout read")
	}
}
