package rules1e

import "testing"

// TestCasterLevelRules: each game's CasterLevelRule as its GetCasterLevel.
func TestCasterLevelRules(t *testing.T) {
	type lv struct{ cleric, paladin, ranger, mage Level }
	cases := []struct {
		name  string
		r     CasterLevelRule
		l     lv
		class SpellClass
		want  int
	}{
		{"por none", CasterLevelRule{}, lv{}, SpellClassCleric, 0},
		{"por power", CasterLevelRule{}, lv{}, SpellClassPower, innateCasterLevel},
		{"por paladin", CasterLevelRule{}, lv{paladin: 12}, SpellClassCleric, 0},
		{"cotab none", CotabCasterLevel, lv{}, SpellClassPower, itemCasterLevel},
		{"cotab paladin", CotabCasterLevel, lv{paladin: 12}, SpellClassCleric, 4},
		{"sotsb paladin", SotsbCasterLevel, lv{paladin: 12, mage: 1}, SpellClassCleric, 4}, // ruled: paladin level - 8 in every game
		{"sotsb paladin only", SotsbCasterLevel, lv{paladin: 12}, SpellClassCleric, itemCasterLevel},
		{"pod ranger druid", PodCasterLevel, lv{ranger: 10}, SpellClassDruid, 3},
		{"pod ranger mage", PodCasterLevel, lv{ranger: 10, mage: 1}, SpellClassMage, 2},
	}
	for _, k := range cases {
		tb := POR
		tb.CasterLevel = k.r
		tb.Spells = [][]SpellProp{make([]SpellProp, len(POR.Spells[1])), make([]SpellProp, len(POR.Spells[1]))}
		tb.Spells[1][0] = SpellProp(k.class)
		c := newChar("C", 0)
		c.Levels[ClsCleric], c.Levels[ClsPaladin], c.Levels[ClsRanger], c.Levels[ClsMage] = k.l.cleric, k.l.paladin, k.l.ranger, k.l.mage
		e := &RuleState{Tables: &tb, Party: []*Character{c}}
		if got := e.casterLevel(1); got != k.want {
			t.Errorf("%s: %d, want %d", k.name, got, k.want)
		}
	}
}

// TestEffectiveLevel: the level a class counts at while the earlier class is active, by the game's rule (the
// scripts' SelMageLevel: cotab-re SpecialReadZ1 $C9 adds, pod's higher), and its own when it is not.
func TestEffectiveLevel(t *testing.T) {
	c := &Character{}
	c.Race = RaceHuman
	c.Levels[ClsMage], c.OldLevels[ClsMage] = 3, 2
	for _, tc := range []struct {
		name string
		r    CasterLevelRule
		want int8
	}{{"cotab adds", CotabCasterLevel, 5}, {"pod takes the higher", PodCasterLevel, 3}, {"por ignores the old", CasterLevelRule{}, 3}} {
		if got := effectiveLevel(tc.r, c, ClsMage); got != tc.want {
			t.Errorf("%s: %d, want %d", tc.name, got, tc.want)
		}
	}
	c.MulticlassLevel = 3 // the earlier class is not active yet
	if got := effectiveLevel(CotabCasterLevel, c, ClsMage); got != 3 {
		t.Errorf("earlier class inactive: %d, want 3", got)
	}
}
