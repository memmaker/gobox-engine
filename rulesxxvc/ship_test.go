package rulesxxvc

import (
	"testing"

	"gobox/rules"
)

// Seed 1: d100 57, Random(100) 56 (outside Go). By hand from shipcbt.pas.
func TestShipArithmetic(t *testing.T) {
	r := rules.RNG{State: 1}
	if got := HitSystem(&r); got != 3 { // 57-28-9-9-13 = -2: life support
		t.Errorf("HitSystem = %d, want 3", got)
	}
	r = rules.RNG{State: 1}
	if ShipSkillCheck(&r, 20, 5, rules.Easy) { // 50 > 56
		t.Error("Easy 25 made")
	}
	r = rules.RNG{State: 1}
	if !ShipSkillCheck(&r, 20, 5, rules.VeryEasy) { // 100 > 56
		t.Error("VeryEasy 25 failed")
	}
	crew := Shooter{Crew: true, BaseCombat: 10, Dex: 16, RocketJock: true, Level: 3, PartyShip: true,
		TargetBonus: 2, Range: 5, TargetManeuver: 2, TargetArmorClass: 30}
	if got := TargetValue(crew); got != 14 { // 10+1+3+3+2-5+2 = 16; 60-30-16
		t.Errorf("TargetValue = %d, want 14", got)
	}
	if got := TargetValue(Shooter{TargetArmorClass: 200}); got != 20 { // -140 as ShortInt is 116
		t.Errorf("TargetValue wrap = %d", got)
	}
	r = rules.RNG{State: 1}
	if !CrewLost(&r, 120, 1) || r.State == 1 { // 60 > 56
		t.Error("CrewLost")
	}
	r = rules.RNG{State: 1}
	if CrewLost(&r, 120, 0) || r.State == 1 {
		t.Error("no crew: rolls, loses none")
	}
}
