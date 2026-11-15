package rulesxxvc

import (
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// Seed 1 under Turbo Pascal's Random, computed outside Go: d6s 5, 6, 5; d10
// then d127: 7, 42; Random(100)s 56, 29. Wanted values by hand from
// cbtacts.pas CalcWeaponDamage and BeginRound.
func TestWeaponDamage(t *testing.T) {
	for _, c := range []struct {
		d             rules.Dice
		backstab, rog bool
		level, want   int
	}{
		{rules.Dice{N: 2, Sides: 6, Bonus: 1}, false, false, 1, 12},
		{rules.Dice{N: 2, Sides: 6, Bonus: 1}, true, true, 5, 36}, // 12 * (12 div 4)
		{rules.Dice{N: 2, Sides: 6, Bonus: 1}, true, false, 5, 24},
		{rules.Dice{N: 1, Sides: 6, Bonus: -9}, false, false, 1, 0},
		{rules.Dice{N: 3, Sides: 6}, true, true, 61, 16}, // 16*17 = 272 in a Byte
	} {
		if got := WeaponDamage(&rules.RNG{State: 1}, c.d, c.backstab, c.rog, c.level); got != c.want {
			t.Errorf("%+v: %d, want %d", c, got, c.want)
		}
	}
}

func TestShielded(t *testing.T) {
	r := rules.RNG{State: 1}
	if Shielded(&r, xxvc.Rocket, Defences{ECM: true}) { // 50 > 56 fails
		t.Error("ECM alone stopped it")
	}
	r = rules.RNG{State: 1}
	if !Shielded(&r, xxvc.Rocket, Defences{ECM: true, Fields: true}) || r.State == 1 { // then 75 > 29
		t.Error("fields did not stop it")
	}
	r = rules.RNG{State: 1}
	if !Shielded(&r, xxvc.Rocket, Defences{SawChaff: true, ECM: true}) || r.State != 1 {
		t.Error("chaff must stop it without a roll")
	}
	if !Shielded(&r, xxvc.Laser, Defences{SawMist: true}) || Shielded(&r, xxvc.Laser, Defences{Body: true}) ||
		!Shielded(&r, xxvc.Microwave, Defences{Body: true}) || Shielded(&r, xxvc.Melee, Defences{SawChaff: true}) {
		t.Error("laser, microwave, melee")
	}
}

func TestInitiative(t *testing.T) {
	for _, c := range []struct {
		dex           int
		surprised     bool
		want, wantTie int
	}{
		{3, false, 5, 42}, // 7 - 3 + 1
		{3, true, 0, 42},  // -1: 0
		{1, false, 2, 42}, // 7 - 6 + 1
		{22, true, 6, 42}, // 12 - 6
	} {
		r := rules.RNG{State: 1}
		if i, tie := Initiative(&r, true, c.dex, c.surprised); i != c.want || tie != c.wantTie {
			t.Errorf("%+v: %d %d", c, i, tie)
		}
	}
	r := rules.RNG{State: 1}
	if i, _ := Initiative(&r, false, 10, false); i != 0 || r.State != 1 {
		t.Error("disabled rolls")
	}
}
