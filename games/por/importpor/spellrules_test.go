package importpor

import (
	"gobox/pool"
	"path/filepath"
	"reflect"
	"rules1e"
	"strings"
	"testing"
)

// TestSpellRulesAligned: every spell's rules start with its own name.
func TestSpellRulesAligned(t *testing.T) {
	for id := 1; id < len(rules1e.SpellNames); id++ {
		if rules1e.SpellNames[id] == "" && SpellRules[id] == "" {
			continue // a removed spell (56 Restoration)
		}
		if !strings.HasPrefix(strings.ToLower(SpellRules[id]), strings.ToLower(rules1e.SpellNames[id])+":") {
			t.Errorf("spell %d %q: rules %q", id, rules1e.SpellNames[id], SpellRules[id])
		}
	}
}

// TestGameDataFile: the data written is what pool reads back, POR's own values.
func TestGameDataFile(t *testing.T) {
	dir := t.TempDir()
	if err := pool.WriteGameData(filepath.Join(dir, pool.DataDir), gameData()); err != nil {
		t.Fatal(err)
	}
	got, err := pool.ReadGameData(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := gameData()
	if got.ID != "por" || len(got.QuickParty) != 6 || !reflect.DeepEqual(got.Rests, want.Rests) || !reflect.DeepEqual(got.VerbOrder, want.VerbOrder) || !reflect.DeepEqual(got.SpellRules, want.SpellRules) || got.DoorInk != want.DoorInk || !reflect.DeepEqual(got.Logo, want.Logo) {
		t.Errorf("read back %+v", got)
	}
}
