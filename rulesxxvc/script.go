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
