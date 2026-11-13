package rules1e

import (
	"slices"

	"gobox/rules"
)

// deathMargin: a hit more than HP+10 kills outright ("dies.", ovr08:20CC).
const deathMargin = 10

// scriptDamage = por opDamage (ovr04:2400) under ruling 3, cotab's DAMAGE
// (O03C:2958) for all games: a made save spares unless SaveHurts; one
// damage roll for the flagged forms, one per member for the random one.
func (r *Rules) scriptDamage(a rules.ScriptDamage) rules.Outcome {
	s := r.S
	if i := slices.Index(s.Party, r.chars[a.Selected]); i >= 0 {
		s.Who = i // the selected member is the actor (e.as(e.sel))
	}
	roll := func() int { return int(byte(s.Dice(int(a.Dice), int(a.Sides)))) + int(a.Bonus) }
	dmg := roll()
	pick := 0
	if !a.All {
		pick = s.Dice(1, int(byte(s.Game.PartyCount)))
	}
	member := func(k int) *Character { // k is 1-based; the walk stops at the end
		return s.Party[max(1, min(k, len(s.Party)))-1]
	}
	var o rules.Outcome
	hit := func(c *Character) {
		if c.Status == StatusDead { // ovr08:20CC skips the dead alone; ApplyDamage leaves a statue whole
			return
		}
		o.Effects = append(o.Effects, rules.Hurt{Unit: r.ID(c), Damage: int32(dmg), Dies: int(c.HP)+deathMargin < dmg})
		s.ApplyDamage(c, dmg&0xFF) // 00BA:2228
	}
	// spared: c saves and a made save spares (the save is rolled even when it hurts anyway).
	spared := func(c *Character, save rules.Save) bool {
		if save == "" {
			return false
		}
		return s.SavingThrow(c, slices.Index(porSave[:], save), int(a.SaveBonus)) && !a.SaveHurts
	}
	switch {
	case len(s.Party) == 0:
	case a.Random:
		for range a.Count {
			if c := member(s.Dice(1, int(byte(s.Game.PartyCount)))); s.RollHit(c, int(a.HitBonus)) { // ovr25:0C4F
				hit(c)
			}
			dmg = roll()
		}
	case a.All:
		for _, c := range s.Party {
			if !spared(c, a.Save) {
				hit(c)
			}
		}
	case a.BySelected:
		if c := r.chars[a.Selected]; c != nil && !spared(c, a.Save) {
			hit(c)
		}
	default:
		if c := member(pick); !spared(c, a.Save) {
			hit(c)
		}
	}
	return o
}
