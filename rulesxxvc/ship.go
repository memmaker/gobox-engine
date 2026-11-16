package rulesxxvc

import (
	"fmt"
	"gobox/rules"
	"gobox/rules/xxvc"
)

// ShipSkillCheck is shipcbt.pas ShipCombat's SkillCheck once the value is
// known (the crew's skill value, or the ship's own systemSkill or
// alternateSkill): value+bonus in a Byte, scaled by h, a percentile roll.
// BR1 = BR2.
func ShipSkillCheck(r *rules.RNG, value, bonus int, h rules.Hardness) bool {
	v := uint8(value + bonus)
	switch h {
	case rules.VeryEasy:
		v <<= 2
	case rules.Easy:
		v <<= 1
	case rules.Hard:
		v /= 2
	case rules.VeryHard:
		v /= 4
	}
	return PercentileRoll(r, int(v))
}

// Shooter is who aims a mount: a crew member (Crew), else the ship's systemValue.
type Shooter struct {
	Crew              bool
	BaseCombat, Dex   int
	RocketJock        bool
	Level             int
	SystemValue       int // no crew
	PartyShip         bool
	TargetBonus       int // shipWeapons[kind].targetBonus
	ControlDown       bool
	TargetControlDown bool
	TargetManeuver    int
	TargetArmorClass  uint8
	Range             int // shipCombatRange
}

// TargetValue is shipcbt.pas TargetValue: the d20 a shot needs, at most 20
// (ShortInt arithmetic; BR1 = BR2).
func TargetValue(s Shooter) int {
	var needed int8
	if s.Crew {
		needed = int8(s.BaseCombat + DexMissileAdjust(s.Dex))
		if s.RocketJock {
			needed += int8(s.Level)
		}
	} else {
		needed = int8(s.SystemValue)
	}
	if s.PartyShip {
		needed += 3
	}
	needed += int8(s.TargetBonus)
	if s.Range > 4 {
		needed -= 5
	} else if s.Range > 2 {
		needed -= 2
	}
	if s.ControlDown {
		needed -= 5
	}
	if s.TargetControlDown {
		needed += 3
	}
	needed += int8(s.TargetManeuver)
	needed = int8(60 - int(s.TargetArmorClass) - int(needed))
	if needed > 20 {
		needed = 20
	}
	return int(needed)
}

// systemHitChances is gdconsts.pas's percent per system, weapons last (BR1 = BR2).
var systemHitChances = [7]int8{28, 9, 9, 13, 10, 18, 13}

// WeaponsHit is HitSystem's index for the weapon mounts (C_Sys.WEAPONS).
const WeaponsHit = 6

// HitSystem is shipcbt.pas DispatchWeaponHit's pick for ANY_SYSTEM: a d100
// down the hit chances; 0..5 a system (xxvc.ShipSystems' order), 6 the mounts.
func HitSystem(r *rules.RNG) int {
	roll := int8(rules.RNGDice(r, 1, 100))
	for i, c := range systemHitChances {
		if roll -= c; roll <= 0 {
			return i
		}
	}
	return WeaponsHit // unreachable: the chances sum to 100
}

// CrewLost is DamageShip's tail: a crew member dies on PercentileRoll(value
// shr 1) when the ship has crew; value is the damage dealt (0 after life
// support failed).
func CrewLost(r *rules.RNG, value, crew int) bool {
	return PercentileRoll(r, int(uint8(uint16(value)>>1))) && crew > 0 // PercentileRoll takes a Byte
}

// shipCell is the cell a ShipCellFact names of the party's ship (ok false: it has no such cell). The word cells
// come back as p, the byte cells as b.
func shipCell(s *xxvc.ShipRecord, f rules.ShipCellFact) (p *int16, b *uint8, ok bool) {
	switch {
	case f.Part == rules.ShipSystemStatus && f.Index >= 0 && f.Index < len(s.Status):
		return &s.Status[f.Index], nil, true
	case f.Part == rules.ShipCrew:
		return &s.Crew, nil, true
	case f.Index < 0 || f.Index >= len(s.Mounts):
		return nil, nil, false
	}
	m := &s.Mounts[f.Index]
	switch f.Part {
	case rules.ShipMountStatus:
		return nil, &m.Status, true
	case rules.ShipMountAmmo:
		return nil, &m.Ammo, true
	case rules.ShipMountReloads:
		return nil, &m.Reloads, true
	}
	return nil, nil, false
}

// foeCell is the cell a ShipCellFact names of the last foe's summary (zone0's savedFoe, the Area[620..642] copy).
func foeCell(s *savedFoe, f rules.ShipCellFact) (p *int16, b *uint8, ok bool) {
	switch {
	case f.Part == rules.ShipSystemStatus && f.Index >= 0 && f.Index < len(s.Status):
		return &s.Status[f.Index], nil, true
	case f.Part == rules.ShipCrew:
		return &s.Crew, nil, true
	case f.Part == rules.ShipWeaponCount:
		return nil, &s.WeaponCount, true
	}
	return nil, nil, false
}

// cell is the address of the number f names.
func (r *Rules) cell(f rules.ShipCellFact) (*int16, *uint8, bool) {
	if f.Foe {
		return foeCell(&r.foeSaved, f)
	}
	r.partyShip()
	return shipCell(r.ship, f)
}

// shipNumber is a number of the party's or the foe's ship (the scripts' reads of Area[512..642]).
func (r *Rules) shipNumber(f rules.ShipCellFact) (int32, error) {
	p, b, ok := r.cell(f)
	switch {
	case !ok:
		return 0, fmt.Errorf("rulesxxvc: no ship cell %+v", f)
	case p != nil && f.Part == rules.ShipCrew:
		return int32(uint8(*p)), nil
	case p != nil:
		return int32(*p), nil
	}
	return int32(*b), nil
}

// setShipCell is a script's write of a ship cell, stored as the record's width (a word or a byte); the crew's
// byte replaces the word's low byte.
func (r *Rules) setShipCell(f rules.ShipCellFact, v rules.Value) error {
	n, isNumber := v.(rules.Number)
	p, b, ok := r.cell(f)
	switch {
	case !ok || !isNumber:
		return fmt.Errorf("rulesxxvc: no ship cell %+v to set to %v", f, v)
	case p != nil && f.Part == rules.ShipCrew:
		*p = int16(uint16(*p)&0xFF00 | uint16(uint8(n)))
	case p != nil:
		*p = int16(n)
	default:
		*b = uint8(n)
	}
	return nil
}
