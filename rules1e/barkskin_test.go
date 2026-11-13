package rules1e

import "testing"

// TestBarkskinRuled: base AC 6, one better per 4 caster levels, no better
// than worn armour gives; +1 on saves but against magic; gone at removal.
func TestBarkskinRuled(t *testing.T) {
	tb := POR
	var err error
	if tb.Affects, err = withVariants(tb.Affects, SotsbAffectVariants); err != nil {
		t.Fatal(err)
	}
	e := &RuleState{Tables: &tb}
	for _, k := range []struct {
		lvl  byte
		base byte // armour on the 60-minus-AC scale
		want ArmourClass
	}{{1, AcTen, barkskinAC}, {4, AcTen, barkskinAC + 1}, {8, AcTen, barkskinAC + 2}, {1, AcTen + 7, AcTen + 7}} {
		c := &Character{}
		c.BaseAC = k.base
		c.Stats[StatDex] = 10
		e.AddAffect(c, affBarkskin, 5, k.lvl, barkskinCall)
		e.ApplyItems(c)
		if c.AC[0] != k.want {
			t.Errorf("level %d base %d: AC %d, want %d", k.lvl, k.base, c.AC[0], k.want)
		}
		e.RemoveAffect(c, affBarkskin, nil)
		if c.AC[0] != ArmourClass(k.base) || len(c.Affects) != 0 {
			t.Errorf("level %d: after removal AC %d, affects %d", k.lvl, c.AC[0], len(c.Affects))
		}
	}
	c := &Character{}
	a := &Affect{AffectRecord: AffectRecord{Type: affBarkskin}}
	for _, k := range []struct {
		dmgType, saveType byte
		want              byte
	}{{0, SaveVsParalysis, 6}, {DmgFire, saveVsBreath, 6}, {DmgFire | DmgMagic, saveVsBreath, 5}, {0, SaveVsSpell, 5}} {
		e.Rv.Save, e.Rv.DmgType, e.Rv.SaveType = 5, k.dmgType, k.saveType
		e.barkskinRuled(c, a, 0)
		if e.Rv.Save != k.want {
			t.Errorf("dmg %d save %d: %d, want %d", k.dmgType, k.saveType, e.Rv.Save, k.want)
		}
	}
}
