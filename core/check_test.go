package core

import (
	"strings"
	"testing"
)

func TestMismatches(t *testing.T) {
	n := Needs{Spells: []SpellID{"knock"}, Facts: []Fact{SkillFact{"climb"}, FactStatus}, Features: []Feature{FeatureSpaceCombat}, Monsters: []MonsterID{"hobgoblin"}}
	got := Mismatches(n, Catalog{Spells: []SpellID{"knock"}, Facts: []Fact{FactStatus}})
	if len(got) != 3 || !strings.Contains(got[0], "climb") || !strings.Contains(got[1], "space-combat") || !strings.Contains(got[2], "hobgoblin") {
		t.Fatal(got)
	}
	if got := Mismatches(n, Catalog{Spells: n.Spells, Facts: n.Facts, Features: n.Features, Monsters: n.Monsters}); got != nil {
		t.Fatal(got)
	}
}

// TestRNGIsTurboPascals: the first values of gobox/rules.Random from seed 1.
func TestRNGIsTurboPascals(t *testing.T) {
	r := RNG{State: 1}
	r.N(0)
	if r.State != 0x08088405+1 {
		t.Fatal(r.State)
	}
	got := r.N(100)
	if want := int(uint16(r.State>>16) % 100); got != want {
		t.Fatal(got, want)
	}
}
