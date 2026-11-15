package rulesxxvc

import (
	"fmt"
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// InflictDamage is charhlp.pas InflictDamage: hit points down to 0; the
// overkill decides the state: over 29 dead, over 9 comatose, else dying,
// exactly 0 unconscious; any of these disables (and in a fight ends the
// initiative, the caller's). It answers whether c went down.
func InflictDamage(c *Character, damage int) bool {
	overkill, remaining := 0, 0
	if c.HP >= damage {
		remaining = c.HP - damage
	} else {
		overkill = damage - c.HP
	}
	c.HP = remaining
	switch {
	case overkill > 29:
		c.Status = StDead
	case overkill > 9:
		c.Status = StComatose
	case overkill > 0:
		c.Status = StDying
	case remaining == 0:
		c.Status = StUnconscious
	default:
		return false
	}
	c.Enabled, c.HP = false, 0
	return true
}

// attack is cbtacts.pas MakeAttack and RollAttacks of a with its attacks left
// this round (phase 2 from scale, then phase 1 from baseScale) against d:
// NoteAttackDirection, the defender turning to face a first or second
// attack, a rear attack (IsRearAttack) on the rear armour, a backstab
// (IsBackstab) or from behind, a rogue behind +2, ApplyRangePenalty; each
// AttackDie that AttackHits, or on a helpless d, deals WeaponDamage unless
// Shielded (InflictDamage), or with the Sonic Stunner stuns unless saved,
// until d is down. With no attacks left the actions end.
// Lines (WalkLine) give the blocked-terrain +2 and chaff/mist shielding; a
// ranged weapon spends its shots.
// The effects: BLIND (attack roll, armour, save), the defender's
// WEAPON_DAMAGE_DEFENDER immunities before the shielding, the attacker's
// COND_HIT_PHASE_2 poisons, the stun through AddEffectWithMessage.
// ponytail: projectiles are not drawn; distance is the king's-move count
// (CalcDistance's line distance not used).
func (r *Rules) attack(f rules.Fight, aid, did rules.CharID, rng *rules.RNG) (rules.Outcome, error) {
	a, d, s := r.chars[aid], r.chars[did], r.fight[aid]
	if a == nil || d == nil || s == nil {
		return rules.Outcome{}, fmt.Errorf("Attack %d on %d: not in the fight", aid, did)
	}
	var o rules.Outcome
	ds := r.fight[did]
	aat, _ := unitAt(f, aid)
	dat, _ := unitAt(f, did)
	facing := unitFacing(f, did)
	if ds != nil {
		ds.attacksDrawn++ // NoteAttackDirection
		if ds.attacksDrawn < 2 {
			facing = rules.Facing(rules.DirectionTo(dat.X, dat.Y, aat.X, aat.Y))
			o.Effects = append(o.Effects, rules.Turned{Unit: unitOf(did), Facing: facing})
		}
	}
	toward := rules.DirectionTo(aat.X, aat.Y, dat.X, dat.Y)
	_, wt, wty, armed := r.readied(a, xxvc.SlotWeapon)
	weaponRange := r.readiedWeaponRange(a)
	rear, backstab := false, false
	if ds != nil && ds.attacksDrawn > 1 && unitSize(f, did) <= 1 {
		switch int(facing) - toward { // a ShortInt: -7 is not behind
		case -1, 0, 1, 7:
			rear = true
		}
	}
	if rear && weaponRange < 2 && int(facing) == toward { // IsBackstab
		h := rules.Normal
		if a.Career == xxvc.Rogue {
			h = rules.Easy
		}
		g1, _ := SkillCheck([]int8{skillValue(a, moveSilentlySkill)}, h, rng)
		backstab = g1 >= rules.Made
		if !backstab {
			g2, _ := SkillCheck([]int8{skillValue(a, distractSkill)}, h, rng)
			backstab = g2 >= rules.Made
		}
	}
	r.RecalcStats(d) // then DEFENDER_ARMOR_CLASS: a blind defender is 4 easier
	if hasEffect(d, EffBlind) {
		d.ArmorClass[0] -= blindPenalty
		d.ArmorClass[1] -= blindPenalty
	}
	armor := ArmorValue(d.ArmorClass[0], d.ArmorClass[1], rear, false, a.Career == xxvc.Rogue)
	_, _, _, forth := walkLine(f.Board, aat, dat, 255, true, false) // ApplyRangePenalty's line
	armor += RangePenalty(weaponRange > 1 && s.guarding, forth.blocked, distance(aat, dat), weaponRange)
	// CalcWeaponDamage's line from the defender back, and the attacker's cell
	_, _, _, back := walkLine(f.Board, dat, aat, 255, true, false)
	switch terrainAt(f.Board, aat) {
	case chaffTerrain:
		back.chaff = true
	case aerosolMistTerrain:
		back.mist = true
	}
	mod := Thac0Modifier(a.team(), difficulty(r.options.Difficulty), int(f.Terms.PartyHit), int(f.Terms.MonsterHit))
	stunner := armed && wt.Type == sonicStunnerType
	hit := false
	s.hasAttacked = true // RollAttacks
	phase1 := 0          // attackCounts[1]
	for phase := 1; phase >= 0 && d.Enabled; phase-- {
		for s.attacks[phase] > 0 && d.Enabled {
			s.attacks[phase]--
			if phase == 0 {
				phase1++
			}
			roll, missed := AttackDie(rng)
			if missed {
				roll = -1
			} else if hasEffect(a, EffBlind) { // ATTACK_ROLL_ATTACKER
				roll -= blindPenalty
			}
			if !AttackHits(roll, a.AttackBonus, mod, armor) && !helpless(d) {
				continue
			}
			if stunner {
				// AddEffectWithMessage's minutes are rolled as its argument, saved or not.
				saved, minutes := rollSave(d, 0, 0, rng), rules.RNGDice(rng, 1, 4)+4
				if r.addEffect(d, a, Effect{ID: EffStunned, Minutes: minutes}, saved) {
					o.Effects = append(o.Effects, rules.Narrated{Text: d.Name + " is stunned"})
				}
				continue
			}
			hit = true
			dmg := WeaponDamage(rng, a.Damage[phase], backstab, a.Career == xxvc.Rogue, a.Level)
			dmg, dodged := defenderDamage(d, dmg, armed, wty.WeaponClass, distance(aat, dat), rng)
			if dodged {
				o.Effects = append(o.Effects, rules.Narrated{Text: d.Name + " dodges it"})
			}
			if armed && Shielded(rng, wty.WeaponClass, Defences{SawChaff: back.chaff, SawMist: back.mist, ECM: d.Armor.ECM, Fields: d.Armor.Fields, Body: d.Armor.Body}) {
				dmg = 0
			}
			o.Effects = append(o.Effects, rules.Attacked{From: unitOf(aid), To: unitOf(did), Hit: true, Damage: int32(dmg)})
			if InflictDamage(d, dmg) {
				if ds != nil {
					ds.initiative = 0
				}
				o.Effects = append(o.Effects, rules.Fell{Unit: unitOf(did)})
			}
			if d.Enabled && dmg > 0 && phase == 1 { // COND_HIT_PHASE_2
				o.Effects = append(o.Effects, r.poison(a, d, did, ds, rng)...)
			}
		}
	}
	if !hit {
		o.Effects = append(o.Effects, rules.Attacked{From: unitOf(aid), To: unitOf(did)})
	}
	// MakeAttack: a ranged weapon spends a shot per phase-1 attack; a
	// consumable one used up goes; then RecalcStats.
	if i := r.readiedIndex(a, xxvc.SlotWeapon); armed && i >= 0 && wty.WeaponClass > xxvc.Melee {
		a.Items[i].Ref.Count -= int32(phase1)
		if a.Items[i].Ref.Count <= 0 && wty.Consumable {
			a.Items = slices.Delete(a.Items, i, i+1)
		}
	}
	r.RecalcStats(a)
	if s.attacks[0] == 0 && s.attacks[1] == 0 || !a.Enabled {
		endActions(s)
	}
	return o, nil
}

// The C_Skill numbers IsBackstab rolls.
var (
	moveSilentlySkill, _ = skillNo("move-silently")
	distractSkill, _     = skillNo("distract") // CbtActs.SKILL_DISTRACT, 78
)

func unitFacing(f rules.Fight, id rules.CharID) rules.Facing {
	for _, u := range f.Board.Units {
		if u.ID == id {
			return u.Facing
		}
	}
	return 0
}

func unitSize(f rules.Fight, id rules.CharID) int32 {
	for _, u := range f.Board.Units {
		if u.ID == id {
			return u.Size
		}
	}
	return 0
}

// helpless is CharHlp.IsHelpless: the HELPLESS effect.
func helpless(c *Character) bool { return hasEffect(c, EffHelpless) }

// attacksLeft is BeginRound's and RecalcAttacksLeft's counts: rate plus the
// round's odd bit, halved.
func attacksLeft(rate int, round int32) int { return (rate + int(round&1)) >> 1 }
