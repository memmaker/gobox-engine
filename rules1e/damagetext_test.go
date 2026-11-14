package rules1e

import "testing"

// A monster's bite lives in hand 1 (giant snake MON5 #60: dice 0,3 sides 0,6 bonus 0,2).
func TestDamageTextSecondHand(t *testing.T) {
	c := &Character{}
	c.Dice, c.Sides, c.Bonus = [2]Count{0, 3}, [2]Count{0, 6}, [2]Modifier{0, 2}
	if got := DamageText(c); got != "3d6+2" {
		t.Errorf("DamageText = %q, want 3d6+2", got)
	}
}
