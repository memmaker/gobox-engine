package rulesxxvc

import (
	"testing"

	"gobox/rules"
)

// By hand from fxengine.pas AttackRoll, cbtacts.pas ProjectileRangePenalty,
// ApplyRangePenalty, RollAttacks (BR1 and BR2 alike).
func TestAttack(t *testing.T) {
	r := rules.RNG{State: 1} // the die is Random(20)+1
	want := rules.RNG{State: 1}
	d := want.N(20) + 1
	if roll, missed := AttackDie(&r); roll != d || missed != (d == 1) {
		t.Errorf("AttackDie = %d %v, die %d", roll, missed, d)
	}
	for _, c := range []struct {
		team bool
		d    Difficulty
		want int
	}{{true, Novice, 5}, {true, Cadet, 4}, {true, Trooper, 3}, {true, Veteran, 2}, {true, Ace, 1}, {false, Novice, -1}} {
		if got := Thac0Modifier(c.team, c.d, 3, -1); got != c.want {
			t.Errorf("Thac0Modifier(%v, %d) = %d", c.team, c.d, got)
		}
	}
	if !AttackHits(naturalAttackMax, 0, -5, 60) || AttackHits(10, 2, 0, 13) || !AttackHits(10, 3, 0, 13) || AttackHits(-1, 50, 0, 0) {
		t.Error("AttackHits")
	}
	for _, c := range [][3]int{{9, 1, 0}, {7, 8, 5}, {6, 8, 2}, {5, 8, 2}, {4, 8, 0}, {7, 9, 5}, {6, 9, 2}} {
		if got := ProjectileRangePenalty(c[0], c[1]); got != c[2] {
			t.Errorf("ProjectileRangePenalty(%d, %d) = %d, want %d", c[0], c[1], got, c[2])
		}
	}
	if RangePenalty(true, true, 9, 8) != 5 || RangePenalty(false, true, 7, 8) != 7 || RangePenalty(false, false, 2, 8) != 0 {
		t.Error("RangePenalty")
	}
	if ArmorValue(4, 6, false, false, true) != 4 || ArmorValue(4, 6, true, false, false) != 6 ||
		ArmorValue(4, 6, true, true, true) != 8 || ArmorValue(4, 6, false, true, true) != 6 {
		t.Error("ArmorValue")
	}
}

func TestSave(t *testing.T) {
	for seed := uint32(0); seed < 200; seed++ {
		r, w := rules.RNG{State: seed}, rules.RNG{State: seed}
		d := w.N(20) + 1
		roll, made, fixed := SaveDie(&r, -3)
		switch {
		case d == 1 && (made || !fixed), d == 20 && (!made || !fixed):
			t.Fatalf("natural %d: made %v fixed %v", d, made, fixed)
		case d > 1 && d < 20 && (fixed || roll != int(uint8(d-3))):
			t.Fatalf("die %d: roll %d fixed %v", d, roll, fixed)
		}
	}
	if !SaveMade(12, 12) || SaveMade(13, 12) {
		t.Error("SaveMade")
	}
}
