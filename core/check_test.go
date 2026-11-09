package core

import (
	"strings"
	"testing"
)

func TestMismatches(t *testing.T) {
	n := Needs{Spells: []SpellID{SpellKnock}, Facts: []Fact{SkillFact{SkillClimb}, FactStatus}, Features: []Feature{FeatureSpaceCombat}}
	got := Mismatches(n, Catalog{Spells: []SpellID{SpellKnock}, Facts: []Fact{FactStatus}})
	if len(got) != 2 || !strings.Contains(got[0], "climb") || !strings.Contains(got[1], "space-combat") {
		t.Fatal(got)
	}
	if got := Mismatches(n, Catalog{Spells: n.Spells, Facts: n.Facts, Features: n.Features}); got != nil {
		t.Fatal(got)
	}
}
