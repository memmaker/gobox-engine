package rules1e

import (
	"testing"

	"gobox/rules"
)

// TestThiefItems pins COTAB's items of thievery and free action (O0FE:0AEA).
func TestThiefItems(t *testing.T) {
	skills := func(lvl int, p ItemPowers) [8]int { return ThiefSkills(lvl, RaceHuman, 12, p) }
	low := skills(1, ItemPowers{Thievery: true})
	if low[0] != skills(5, ItemPowers{})[0] || low[1] != skills(7, ItemPowers{})[1] || low[2] != skills(1, ItemPowers{})[2] {
		t.Errorf("thievery at 1 lifts the first two skills' levels only: %v", low)
	}
	high, plain := skills(9, ItemPowers{Thievery: true}), skills(9, ItemPowers{})
	for i := range high {
		if high[i] != plain[i]+thieveryBonus {
			t.Errorf("thievery at 9, skill %d: %d, want %d (the bonus kept across skills)", i, high[i], plain[i]+thieveryBonus)
		}
	}
	if skills(1, ItemPowers{FreeAction: true}) != skills(freeActionLevel, ItemPowers{}) {
		t.Error("free action below level 4 lifts the level, no bonus")
	}
	if free, plain := skills(9, ItemPowers{FreeAction: true}), skills(9, ItemPowers{}); free[3] != plain[3]+freeActionBonus {
		t.Errorf("free action at 9 adds %d: %v", freeActionBonus, free)
	}
}

// TestAffectNames: every handler name POR's Affects uses has a handler (a rules.json with an unknown name would leave the slot empty).
func TestAffectNames(t *testing.T) {
	for typ, name := range POR.Affects {
		if affectHandler[name] == nil {
			t.Errorf("affect %#x: no handler %q", typ, name)
		}
	}
}

// TestConSave pins COTAB's constitution bonus in the stored paralysis save (O0FE:07FB).
func TestConSave(t *testing.T) {
	para := func(r Who) int { return ClassStats(r).Saves[SaveParalysis] }
	lv := map[rules.Class]int{ClassFighter: 1}
	base := para(Who{Levels: lv, Race: RaceHuman, Con: 18})
	if got := para(Who{Levels: lv, Race: RaceDwarf, Con: 18}); got != base-5 {
		t.Errorf("dwarf con 18: %d, want %d", got, base-5)
	}
	if got := para(Who{Levels: lv, Race: RaceHuman, Con: 18, Powers: ItemPowers{SaveBonus: true}}); got != base-5 {
		t.Errorf("save-bonus item con 18: %d, want %d", got, base-5)
	}
	if got := para(Who{Levels: lv, Race: RaceHalfling, Con: 19}); got != base-1 {
		t.Errorf("halfling con 19: %d, want %d (19 is past the low table)", got, base-1)
	}
}

// TestPaladinKnowsClericSpells: a paladin with cleric slots (L9+) knows those levels' cleric spells.
func TestPaladinKnowsClericSpells(t *testing.T) {
	var e RuleState
	c := &Character{}
	c.Levels[ClsPaladin] = 9
	e.RecomputeCharacter(c)
	known := 0
	for k := 1; k <= LastSpell; k++ {
		if c.SpellBook[k-1] != 0 {
			known++
			if rec := POR.Spells[k][:2]; rec[0] != 0 || c.SpellSlot(0, int(int8(rec[1]))) == 0 {
				t.Errorf("spell %d known without a cleric slot of its level", k)
			}
		}
	}
	if known == 0 {
		t.Error("a level 9 paladin knows no cleric spells")
	}
}

// TestLoadDropsConAffect: a loaded record's affect 0x5A goes (its bonus is in the stored save).
func TestLoadDropsConAffect(t *testing.T) {
	var e RuleState
	c := &Character{Affects: []*Affect{{AffectRecord: AffectRecord{Type: affSaveBonusVsType0}}, {AffectRecord: AffectRecord{Type: affSaveBonusVsType42}}}}
	e.AddToParty(c)
	if e.FindAffect(c, affSaveBonusVsType0) != nil || e.FindAffect(c, affSaveBonusVsType42) == nil {
		t.Errorf("affects after load: %v", c.Affects)
	}
}

// ClassStats is ClassCal.RecalcClassStats over POR's tables.
func ClassStats(w Who) Stats { return POR.ClassStats(w) }
