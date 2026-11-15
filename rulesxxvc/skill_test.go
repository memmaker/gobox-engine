package rulesxxvc

import (
	"testing"

	"gobox/rules"
)

// Seed 1's d100s under Turbo Pascal's Random (state*0x08088405+1, high word
// mod 100, +1), computed outside Go: 57, 30, 77. The wanted grades follow
// charhlp.pas SelectSkillCharacter by hand.
func TestSkillCheck(t *testing.T) {
	for _, c := range []struct {
		values []int8
		h      rules.Hardness
		grade  rules.Grade
		who    int
	}{
		{[]int8{50}, rules.Normal, rules.Failed, 0},        // -7, not beyond 50
		{[]int8{50}, rules.Easy, rules.Made, 0},            // 100-57 = 43 < 50
		{[]int8{50}, rules.VeryEasy, rules.MadeWell, 0},    // 200-57 = 143 >= 100
		{[]int8{10}, rules.VeryHard, rules.FailedBadly, 0}, // 2-57: 55 > 2
		// margins -47, 10, -72: the second wins; the bands use the last
		// value 5 (the original's), so 10 >= 2 is core.MadeWell, not core.Made.
		{[]int8{10, 40, 5}, rules.Normal, rules.MadeWell, 1},
		{nil, rules.Normal, rules.FailedBadly, -1},
	} {
		r := rules.RNG{State: 1}
		if g, who := SkillCheck(c.values, c.h, &r); g != c.grade || who != c.who {
			t.Errorf("%v %v: %v %d, want %v %d", c.values, c.h, g, who, c.grade, c.who)
		}
	}
}

func TestSkillValue(t *testing.T) {
	for _, c := range []struct {
		level, ability  int8
		prereq, enabled bool
		want            int8
	}{
		{5, 12, true, true, 17}, {5, 12, false, true, 5}, {0, 12, true, true, 0}, {5, 12, true, false, 0},
	} {
		if got := SkillValue(c.level, c.ability, c.prereq, c.enabled); got != c.want {
			t.Errorf("%+v: %d", c, got)
		}
	}
}

func TestRollDice(t *testing.T) {
	r := rules.RNG{State: 1}
	if got := rules.RNGDice(&r, 2, 100); got != 57+30 {
		t.Errorf("2d100 = %d, want 87", got)
	}
}

// goldbox.RNGDice draws what the old local RollDice drew for every count and
// sides a Byte holds (charhlp.pas RollDice(count, sides: Byte)): all calls
// pass constants up to 127, core.Dice parts or a target count within that.
func TestRNGDiceMatchesOldRollDice(t *testing.T) {
	old := func(r *rules.RNG, count, sides int) int {
		var total uint8
		for range count {
			total += uint8(r.N(sides) + 1)
		}
		return int(total)
	}
	for n := 0; n < 256; n += 5 {
		for s := 0; s < 256; s++ {
			a, b := rules.RNG{State: uint32(n*256 + s)}, rules.RNG{State: uint32(n*256 + s)}
			if x, y := old(&a, n, s), rules.RNGDice(&b, n, s); x != y || a != b {
				t.Fatalf("%dd%d: old %d, RNGDice %d", n, s, x, y)
			}
		}
	}
}
