package rules1e

import "testing"

// TestConHitPoints: a Con change works out max HP again (the hit points
// follow) and regeneration is on from Con 20, off below.
func TestConHitPoints(t *testing.T) {
	e := &RuleState{Tables: &POR}
	c := newChar("HERO", 1)
	c.Levels[ClsFighter], c.HPRoll, c.MaxHP, c.HP = 5, 30, 30, 25
	for _, k := range []struct {
		con     Score
		max, hp byte
		regen   bool
	}{{16, 40, 35, false}, {20, 55, 50, true}, {14, 30, 25, false}} {
		c.Stats[StatCon] = k.con
		e.conHitPoints(c)
		if c.MaxHP != k.max || c.HP != k.hp || (e.FindAffect(c, affRegenerate) != nil) != k.regen {
			t.Errorf("con %d: max %d hp %d regen %v, want %d %d %v", k.con, c.MaxHP, c.HP, e.FindAffect(c, affRegenerate) != nil, k.max, k.hp, k.regen)
		}
	}
}

// TestConRegenOwnOnly: Con below 20 ends only the regeneration Con gave; a
// ring's or a troll's stays.
func TestConRegenOwnOnly(t *testing.T) {
	e := &RuleState{Tables: &POR}
	c := newChar("HERO", 1)
	c.Levels[ClsFighter], c.HPRoll, c.MaxHP, c.HP = 5, 30, 30, 25
	e.AddAffect(c, affRegenerate, 0x3C, 1, 1) // another source's
	c.Stats[StatCon] = 20
	e.conHitPoints(c)
	c.Stats[StatCon] = 14
	e.conHitPoints(c)
	n := 0
	for _, a := range c.Affects {
		if a.Type == affRegenerate {
			n++
			if a.Data == conRegenData {
				t.Error("Con's regeneration stayed")
			}
		}
	}
	if n != 1 {
		t.Errorf("%d regenerations, want the other source's", n)
	}
}
