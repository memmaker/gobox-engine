package rules1e

// conHitDiceCap is each class's last level that rolls a hit die (cleric ..
// monk; CotAB's and SotSB's table, the 1e one of every game).
var conHitDiceCap = [8]Level{10, 15, 10, 10, 11, 12, 11, 19}

// regenerateCon: from this constitution on the character regenerates.
const regenerateCon Score = 20

// conRegenData marks the regeneration Con gives (IsHardier's, conHitPoints'): its
// Data, kept by each tick (regenerate).
const conRegenData byte = 0xFF

// conHitPoints = CotAB conHitPoints (cotab stats.go:227, SotSB's is the same):
// the maximum hit points worked out again for c's constitution: the rolled ones
// plus the constitution bonus of each class level (fighters more), shared out
// over the classes; the hit points follow the change. Regeneration is on from
// regenerateCon on and off below (Felix 2026-11-12).
func (e *RuleState) conHitPoints(c *Character) {
	con := byte(c.Stats[StatCon])
	classes := 0
	oldMax := c.MaxHP
	maxHP := c.HPRoll
	var bonus, multi byte // no game here keeps a multiclass level (stats.MulticlassLevel)
	addLevels := func(k int, lvl byte) {
		if byte(conHitDiceCap[k]) <= lvl {
			lvl = byte(conHitDiceCap[k]) - 1
		}
		if k == ClsRanger && (multi == 0 || byte(c.OldLevels[ClsRanger]) == multi) {
			lvl++
		}
		if k == ClsFighter || k == ClsPaladin || k == ClsRanger {
			switch {
			case con >= 15 && con <= 19:
				bonus += lvl * (con - 14)
			case con == 20:
				bonus += lvl * 5
			case con >= 21 && con <= 23:
				bonus += lvl * 6
			case con >= 24 && con <= 25:
				bonus += lvl * 7
			}
			return
		}
		if con > 15 {
			bonus += lvl * 2
		} else if con == 15 {
			bonus += lvl
		}
	}
	for k := range c.Levels {
		if lvl := byte(c.OldLevels[k]); int8(lvl) > 0 {
			addLevels(k, lvl)
		}
		lvl := byte(c.Levels[k])
		if int8(lvl) <= 0 {
			continue
		}
		classes++
		lvl = min(lvl, byte(conHitDiceCap[k]))
		if lvl > multi {
			addLevels(k, lvl-multi)
		}
	}
	if classes > 0 {
		bonus /= byte(classes)
	}
	maxHP += bonus
	c.MaxHP = maxHP
	switch {
	case maxHP > oldMax:
		c.HP += maxHP - oldMax
	case maxHP < oldMax:
		if lost := oldMax - maxHP; c.HP > lost {
			c.HP -= lost
		} else {
			c.HP = 0
		}
	}
	var a *Affect // the regeneration Con gave; another source's stays as it is
	for _, x := range c.Affects {
		if x.Type == affRegenerate && x.Data == conRegenData {
			a = x
		}
	}
	switch {
	case c.Stats[StatCon] >= regenerateCon && a == nil:
		e.AddAffect(c, affRegenerate, 0x3C, conRegenData, 1) // as IsHardier
	case c.Stats[StatCon] < regenerateCon && a != nil:
		a.Call = 0 // ends without its handler's tick (which heals and adds it again)
		e.RemoveAffect(c, affRegenerate, a)
	}
}
