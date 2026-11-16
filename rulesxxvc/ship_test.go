package rulesxxvc

import (
	"gobox/rules/xxvc"
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

// TestShipCells: the scripts' ship cells (ShipCellFact) read and write the party's ship and the foe's summary
// by the record's width; the crew's byte replaces the word's low byte.
func TestShipCells(t *testing.T) {
	r := &Rules{tables: xxvc.Tables{Ships: map[rules.ShipID]xxvc.ShipRecord{xxvc.PartyShip: {Status: [6]int16{1, 2, 3, 4, 5, 6}}}}}
	fuel := rules.ShipCellFact{Part: rules.ShipSystemStatus, Index: 4}
	if n, err := r.shipNumber(fuel); err != nil || n != 5 {
		t.Fatalf("fuel %d %v", n, err)
	}
	if err := r.SetFact(rules.NoChar, fuel, rules.Number(400)); err != nil {
		t.Fatal(err)
	}
	if n, _ := r.shipNumber(fuel); n != 400 {
		t.Errorf("fuel word %d", n)
	}
	r.foeSaved.Crew = 0x0102
	crew := rules.ShipCellFact{Foe: true, Part: rules.ShipCrew}
	if err := r.SetFact(rules.NoChar, crew, rules.Number(9)); err != nil || r.foeSaved.Crew != 0x0109 {
		t.Errorf("foe crew %#x %v", r.foeSaved.Crew, err)
	}
	if err := r.SetFact(rules.NoChar, rules.ShipCellFact{Part: rules.ShipMountAmmo, Index: 8}, rules.Number(1)); err == nil {
		t.Error("a ninth mount was set")
	}
}
