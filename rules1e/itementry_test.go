package rules1e

import (
	"encoding/json"
	"gobox/rules"
	"slices"
	"testing"
)

// TestItemsEnterConverted: a wand from each path game items enter by
// (Import, Spawn, a treasure block, the library, a kit) holds the unified
// id of its game spell number, shifted bytes decoded; Export writes the
// game's byte back. A path that skips the conversion keeps 5 or 0x52.
func TestItemsEnterConverted(t *testing.T) {
	tb := POR
	tb.SpellIDs = SpellMap{5: 70, 59: 71}
	wand := func(b ItemPower) Item { var it Item; it.Spells[1] = b; return it }
	game := []Item{wand(5), wand(0x52)}
	tb.Library = []*Item{&game[0], &game[1]}
	tb.Start.Kit[kitOther] = game
	s := &RuleState{Tables: &tb, Items: make([]ItemBase, 1)}
	r := &Rules{S: s,
		Monsters:      func(rules.MonsterID) (Monster, bool) { return Monster{Items: game}, true },
		TreasureBlock: func(int) []Item { return game }}
	check := func(path string, got []*Item) {
		t.Helper()
		var v []ItemPower
		for _, it := range got {
			v = append(v, it.Spells[1])
		}
		slices.Sort(v)
		if !slices.Equal(v, []ItemPower{70, 71}) {
			t.Errorf("%s: spells %v, want 70 and 71", path, v)
		}
	}
	b, _ := json.Marshal(CharacterJSON{Items: game})
	id, err := r.Import(b)
	if err != nil {
		t.Fatal(err)
	}
	check("Import", r.chars[id].Items)
	out, err := r.Export(id)
	if err != nil {
		t.Fatal(err)
	}
	var f CharacterJSON
	if err := json.Unmarshal(out, &f); err != nil || f.Items[0].Spells[1] != 5 || f.Items[1].Spells[1] != 0x52 {
		t.Errorf("Export: %v %v", f.Items, err)
	}
	m, err := r.Spawn(1, nil)
	if err != nil {
		t.Fatal(err)
	}
	check("Spawn", r.chars[m].Items)
	if _, err := r.Resolve(nil, rules.AddTreasureBlock{Block: 1}, nil); err != nil {
		t.Fatal(err)
	}
	check("TreasureBlock", s.Treasure)
	s.Treasure = nil
	if err := s.AddLibraryItems([]int32{1, 2}, false); err != nil {
		t.Fatal(err)
	}
	check("Library", s.Treasure)
	c := &Character{}
	if err := s.startKit(c); err != nil {
		t.Fatal(err)
	}
	check("Kit", c.Items)
	if game[0].Spells[1] != 5 || game[1].Spells[1] != 0x52 {
		t.Errorf("the game's records changed: %v", game)
	}
}
