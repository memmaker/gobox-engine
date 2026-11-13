package rules1e

// The adventuring loop: ovr04:31B4/32F8 (game loop and area change),
// ovr15 (command input, movement, doors) and ovr21 (clock).

// PartyHasSkill = ovr15:0236 (n): someone has a positive thief skill
// +0x96+n.
func (e *RuleState) PartyHasSkill(n int) bool {
	for _, c := range e.Party {
		if int8(c.Levels[n]) > 0 {
			return true
		}
	}
	return false
}

// SpellSlot = ovr15:05AB (c, spell): index of the memorized spell, or 0xFF.
func SpellSlot(c *Character, spell SpellID) int {
	for i := 0; i <= 0x14; i++ {
		if c.Memorized[i] == spell {
			return i
		}
	}
	return 0xFF
}

// FindCaster = ovr15:05EE (spell): first member with the spell memorized.
// status okay, as the Pascal's fix (the original took the dead and the sleeping).
func (e *RuleState) FindCaster(spell SpellID) *Character {
	for _, c := range e.Party {
		if c.Status == StatusOkay && SpellSlot(c, SpellID(spell)) != 0xFF {
			return c
		}
	}
	return nil
}

// SpendKnock is Knock cast (PartyCastsKnock's choice): the first okay member
// with it memorized, the slot spent; nil with none.
func (e *RuleState) SpendKnock() *Character {
	c := e.FindCaster(SpKnock)
	if c != nil {
		c.Memorized[SpellSlot(c, SpKnock)] = 0
	}
	return c
}

var clockLimits = [7]int{10, 10, 6, 24, 30, 12, 0x100}

// normalizeClock = ovr21:02F2: carry each unit into the next; a year
// passing ages the party.
func (e *RuleState) normalizeClock(t *[7]int) {
	for i := 0; i <= 6; i++ {
		if t[i] < clockLimits[i] {
			continue
		}
		if i == 6 {
			for _, c := range e.Party {
				c.Age = c.Age + 1
			}
			e.ageCheckDue = true // SotSB camp: the next fight's end sees who dies of old age (oldAgeDeaths; none without a rule)
			continue
		}
		t[i+1]++
		t[i] -= clockLimits[i]
	}
}

// AdvanceClock = ovr21:03D5 (count, unit): add count units to the area
// clock (area+0x18C, 7 words), then run down timed effects.
func (e *RuleState) AdvanceClock(count, unit int) {
	var t [7]int
	for i := range t {
		t[i] = int(e.Area.Clock[i])
	}
	for i := 1; i <= count; i++ {
		t[unit] = int(uint16(t[unit] + 1))
		e.normalizeClock(&t)
	}
	for i := range t {
		e.Area.Clock[i] = uint16(t[i])
	}
	e.tickEffects(count, unit)
}

// tickEffects = ovr21:0000 (count, unit): count down the timed entries in
// each member's effect list (+0x7F) in steps of up to 10 rounds; expired
// ones are removed. While camping only members flagged in DS:471D are
// looked at.
func (e *RuleState) tickEffects(count, unit int) {
	r := e
	if e.BeginFlag != 2 {
		for i := 1; i < len(r.EffTick); i++ { // FillChar(DS:471E, 0x48, 1)
			r.EffTick[i] = 1
		}
	} else {
		any := false
		for i := 1; ; i++ { // repeat .. until: index 1 is always looked at
			if i < len(r.EffTick) && r.EffTick[i] != 0 {
				any = true
			}
			if any || i+1 > int(e.Game.PartyCount) {
				break
			}
		}
		if !any {
			return
		}
	}
	w := count
	for u := unit; u > 1; u-- {
		w = int(uint16(w * clockLimits[u-1]))
	}
	next := func(c *Character, a *Affect) *Affect {
		for i, x := range c.Affects {
			if x == a && i+1 < len(c.Affects) {
				return c.Affects[i+1]
			}
		}
		return nil
	}
	head := func(c *Character) *Affect {
		if len(c.Affects) == 0 {
			return nil
		}
		return c.Affects[0]
	}
	for w > 0 {
		step := w
		if step > 10 {
			step = 10
		}
		for i, c := range e.Party {
			if i+1 >= len(r.EffTick) || r.EffTick[i+1] == 0 {
				continue
			}
			r.EffTick[i+1] = 0
			cur, pred := head(c), head(c) // bp-1B, bp-13
			var last *Affect              // bp-0F: the tail when the walk starts
			for a := cur; a != nil; a = next(c, a) {
				last = a
			}
			for done := false; cur != nil && !done; {
				done = cur == last
				d := int(cur.Time)
				switch {
				case d == 0:
					cur = next(c, cur)
				case step < d:
					cur.Time = uint16(d - step)
					r.EffTick[i+1] = 1
					cur = next(c, cur)
				default:
					after := next(c, cur)
					e.RemoveAffect(c, cur.Type, cur) // ovr25:0026
					// ovr21:0203 quirk: when the node before the removed one
					// is the head, the walk restarts at the head.
					if pred == head(c) {
						cur = head(c)
					} else {
						cur = after
					}
				}
				// ovr21:0239: pred = the node before cur (or cur itself).
				pred = head(c)
				for pred != nil && pred != cur && next(c, pred) != cur {
					pred = next(c, pred)
				}
			}
			for ; cur != nil; cur = next(c, cur) { // ovr21:028E
				if cur.Time > 0 {
					r.EffTick[i+1] = 1
				}
			}
		}
		if w > 10 {
			w -= 10
		} else {
			w = 0
		}
	}
}
