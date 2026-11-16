package rulesxxvc

import (
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// carriesKind is whether the item is of the kind (its template's type, xxvc.ItemKinds): EclFindItem's and
// EclDestroyItems' compare.
func (r *Rules) carriesKind(ci Carried, kind rules.ItemID) bool {
	t, _, ok := r.itemType(ci)
	return ok && xxvc.ItemKinds[t.Type] == kind
}

// takeItems is eclvm.pas EclDestroyItems for each member: the items of the kind go (a readied one is
// recalculated away with it); one each: the first only, and a joinable stack of more than one loses one.
func (r *Rules) takeItems(a rules.TakeItems) {
	for _, id := range a.Whom {
		c := r.chars[id]
		if c == nil {
			continue
		}
		for i := 0; i < len(c.Items); {
			ci := c.Items[i]
			if !r.carriesKind(ci, a.Item) {
				i++
				continue
			}
			if t, _, _ := r.itemType(ci); a.OneEach && !t.NotJoinable && quantity(ci, t) > 1 {
				c.Items[i].Ref.Count--
				break
			}
			c.Items = slices.Delete(c.Items, i, i+1)
			if a.OneEach {
				break
			}
		}
		r.RecalcStats(c)
	}
}

// stashItems is eclvm.pas EclTransferCharacterLists for each member: the source list goes in front of the
// destination's, the source is empty, RecalcStats.
func (r *Rules) stashItems(a rules.StashItems) {
	for _, id := range a.Whom {
		c := r.chars[id]
		if c == nil {
			continue
		}
		if a.Away {
			c.Stashed, c.Items = slices.Concat(c.Items, c.Stashed), nil
		} else {
			c.Items, c.Stashed = slices.Concat(c.Stashed, c.Items), nil
		}
		r.RecalcStats(c)
	}
}

// skillDamage is eclvm.pas EclSkillDamage: each who fails the roll takes the damage (EclInflictDamage: a
// member not okay is passed over; "dies" when the hit points and ten do not cover it).
func (r *Rules) skillDamage(a rules.SkillDamage, rng *rules.RNG) rules.Outcome {
	var o rules.Outcome
	n, known := SkillNo(a.Skill), a.Skill >= 1 && int(a.Skill) <= len(xxvc.Skills) // a number outside the 84 reads no level
	for _, id := range a.Whom {
		c := r.chars[id]
		if c == nil {
			continue
		}
		var value int8
		if known {
			value = skillValue(c, n)
		}
		if g, _ := SkillCheck([]int8{value}, a.Hardness, rng); g >= rules.Made {
			continue
		}
		dmg := int(uint8(rules.RNGDice(rng, int(a.Dice), int(a.Sides)) + int(a.Bonus))) // RollDice and the damage are Bytes
		if c.Status != StOkay {
			continue
		}
		o.Effects = append(o.Effects, rules.Hurt{Unit: id, Damage: int32(dmg), Dies: c.HP+10 < dmg})
		InflictDamage(c, dmg)
	}
	return o
}

// scriptDamage is eclvm.pas EclDamage: the damage is rolled once (a Word: a Byte roll plus a Byte bonus), and then
// the random member's place when the whole party is not meant (memberIndex, rolled before any branch). The flagged
// forms hit the whole party, the selected member or the random one, each with a saving throw of the type named
// unless there is none (Save ""); a made save spares unless SaveHurts. The unflagged form makes Count attacks on
// random members, each rolled against HitBonus (AttackRollNoAttacker) with the damage rolled anew after each.
// EclInflictDamage passes over a member who is not okay and "dies" when the hit points and ten do not cover it.
// The message "party is killed" and its pause are the engine's.
func (r *Rules) scriptDamage(a rules.ScriptDamage, rng *rules.RNG) rules.Outcome {
	var o rules.Outcome
	roll := func() int { return int(byte(rules.RNGDice(rng, int(a.Dice), int(a.Sides)))) + int(a.Bonus) }
	dmg := roll()
	pick := 0
	if !a.All {
		pick = rules.RNGDice(rng, 1, int(r.memberCount))
	}
	member := func(k int) rules.CharID { // 1-based; the original's walk down the list would run off its end, here it stops
		if len(r.party) == 0 {
			return rules.NoChar
		}
		return r.party[max(1, min(k, len(r.party)))-1]
	}
	hurt := func(id rules.CharID) {
		c := r.chars[id]
		if c == nil || c.Status != StOkay {
			return
		}
		o.Effects = append(o.Effects, rules.Hurt{Unit: id, Damage: int32(dmg), Dies: c.HP+10 < dmg})
		InflictDamage(c, dmg)
	}
	// spared: c rolls the save of the type and a made save spares unless it hurts anyway; no type, no roll.
	spared := func(id rules.CharID) bool {
		c := r.chars[id]
		kind := saveIndex(a.Save)
		if a.Save == "" || kind < 0 {
			return false
		}
		return c != nil && rollSave(c, kind, int8(byte(a.SaveBonus)), rng) && !a.SaveHurts
	}
	switch {
	case a.Random:
		for range a.Count {
			// AttackRollNoAttacker: a natural 1 misses, a natural 20 counts 100, else the roll plus the target must
			// pass the defender's armor class; the effects at COND_ATTACK_ROLL_DEF are none.
			id := member(rules.RNGDice(rng, 1, int(r.memberCount)))
			d := rules.RNGDice(rng, 1, 20)
			if c := r.chars[id]; c != nil && d > 1 && (d == 20 || d+int(byte(a.HitBonus)) > c.ArmorClass[0]) {
				hurt(id)
			}
			dmg = roll()
		}
	case a.All:
		for _, id := range r.party {
			if !spared(id) {
				hurt(id)
			}
		}
	case a.BySelected:
		if !spared(a.Selected) {
			hurt(a.Selected)
		}
	default:
		if id := member(pick); !spared(id) {
			hurt(id)
		}
	}
	return o
}

// gainExperience is eclvm.pas EclAwardExperience's award: each of Whom (the party, or the current character) that
// is enabled gains the points (the message is the engine's).
func (r *Rules) gainExperience(a rules.GainExperience) {
	for _, id := range a.Whom {
		if c := r.chars[id]; c != nil && c.Enabled {
			c.Exp += int64(a.Points)
		}
	}
}
