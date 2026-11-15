package rulesxxvc

import (
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

func tables() xxvc.Tables {
	return xxvc.Tables{
		ItemTypes: []xxvc.ItemType{{Slot: xxvc.SlotWeapon, Damage: rules.Dice{N: 1, Sides: 6}, Range: 6}},
		ItemWords: []string{"", "Laser", "Pistol"},
		Items:     map[rules.ItemID]xxvc.ItemTemplate{7: {Type: 0, Words: [3]int32{1, 2, 0}, Plus: 1, Quantity: 1, Value: 50}},
		Monsters:  map[rules.MonsterID]xxvc.Monster{3: {Name: "RAM trooper", HP: 9, Hostile: true, XPAward: 20}},
	}
}

// EclLoadMonster: (d+2)*hp div 4: novice 9 -> 4, trooper 9, ace 9 -> 13; 1 -> 0 -> 1 at novice.
func TestSpawn(t *testing.T) {
	for _, c := range [][3]int{{9, int(Novice), 4}, {9, int(Trooper), 9}, {9, int(Ace), 13}, {1, int(Novice), 1}} {
		if got := SpawnHP(c[0], Difficulty(c[1])); got != c[2] {
			t.Errorf("SpawnHP(%d, %d) = %d", c[0], c[1], got)
		}
	}
	r := New()
	r.Begin(rules.Campaign{Tables: tables()}, rules.Options{})
	id, err := r.Spawn(3, nil)
	if err != nil || r.chars[id].Side != rules.SideMonsters || r.Sheet(id).HP != 9 {
		t.Error(err, r.chars[id])
	}
	if _, err := r.Spawn(4, nil); err == nil {
		t.Error("no monster 4")
	}
	it, err := r.Item(rules.ItemRef{ID: 7})
	if err != nil || it.Kind != rules.KindWeapon || it.Range != 6 || it.Plus != 1 || len(it.NameWords) != 2 || it.NameWords[1] != "Pistol" {
		t.Errorf("%v %+v", err, it)
	}
}
