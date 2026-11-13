package rules1e

import (
	"gobox/core"
	"testing"
)

// TestPaladinPowers pins the lay on hands (2 hp per level, the day's use
// affect set) and the cure (diseases gone, the cured affect set).
func TestPaladinPowers(t *testing.T) {
	s := &RuleState{}
	r := &Rules{S: s}
	pal := &Character{}
	pal.Status, pal.HP, pal.MaxHP = StatusOkay, 30, 30
	pal.Levels[ClsPaladin] = 4
	hurt := &Character{}
	hurt.Status, hurt.HP, hurt.MaxHP = StatusOkay, 5, 30
	s.Party = []*Character{pal, hurt}
	p, h := r.ID(pal), r.ID(hurt)
	out, err := r.Resolve(nil, core.PaladinHeal{Who: p, Target: h}, nil)
	if err != nil || hurt.HP != 13 || len(out.Effects) != 1 || s.FindAffect(pal, AffLaidOnHands) == nil {
		t.Fatalf("heal: hp %d, effects %v, err %v", hurt.HP, out.Effects, err)
	}
	s.AddAffect(hurt, affSick, 100, 0, 0)
	if !r.Check(core.Diseased{Who: h}, nil).Yes {
		t.Fatal("sick is not diseased")
	}
	if _, err := r.Resolve(nil, core.PaladinCure{Who: p, Target: h}, nil); err != nil {
		t.Fatal(err)
	}
	if r.Check(core.Diseased{Who: h}, nil).Yes || s.FindAffect(pal, AffCuredDisease) == nil {
		t.Error("cure left the disease or did not mark the use")
	}
}

// TestMarkImported: a game that flags added characters (rules.json importFlag: COTAB, as from
// Hillsfar) sets the flag the scripts read; one that does not leaves it be.
func TestMarkImported(t *testing.T) {
	for _, flagged := range []bool{true, false} {
		tab := POR
		tab.ImportFlag = flagged
		s := &RuleState{Tables: &tab}
		r := &Rules{S: s}
		c := &Character{}
		id := r.ID(c)
		if _, err := r.Resolve(nil, core.MarkImported{Who: id}, nil); err != nil {
			t.Fatal(err)
		}
		if got := c.ScriptFlag&ImportedFlag != 0; got != flagged {
			t.Errorf("importFlag %v: flag %v", flagged, got)
		}
	}
}
