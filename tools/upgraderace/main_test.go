package main

import (
	"encoding/json"
	"testing"

	"gobox/rules"
	"gobox/rules1e"
)

// TestUpgrade: an old-format monster file (race numbers) becomes the new one
// by the game's numbering, reads as a Sheet, and is then left alone.
func TestUpgrade(t *testing.T) {
	old := []byte(`{"1":{"Record":{"CharName":"A","Race":3},"Items":[]},"2":{"Record":{"CharName":"B","Race":9},"Items":null}}`)
	for _, c := range []struct {
		layout   rules.RecordLayout
		race1    rules.Race
		race2    rules.Race
		monster2 byte
	}{
		{rules1e.Layout285, rules1e.RaceGnome, rules1e.RaceMonster, 9},
		{rules1e.Layout439, rules1e.RaceDwarf, rules1e.RaceMonster, 9},
	} {
		b, changed, err := upgrade(c.layout, old)
		if err != nil || !changed {
			t.Fatalf("layout %d: %v %v", c.layout, changed, err)
		}
		var file map[string]struct{ Record rules1e.Sheet }
		if err := json.Unmarshal(b, &file); err != nil {
			t.Fatal(err)
		}
		if s := file["1"].Record; s.Race != c.race1 || s.CharName != "A" {
			t.Errorf("layout %d: monster 1 is %q %q", c.layout, s.CharName, s.Race)
		}
		if s := file["2"].Record; s.Race != c.race2 || s.MonsterRace != c.monster2 {
			t.Errorf("layout %d: monster 2 is %q %d", c.layout, s.Race, s.MonsterRace)
		}
		if _, changed, err := upgrade(c.layout, b); err != nil || changed {
			t.Errorf("layout %d: second run changed %v (%v)", c.layout, changed, err)
		}
	}
}
