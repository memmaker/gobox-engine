package rules1e

import (
	"gobox/rules"
	"slices"
	"strconv"
)

// GAME.OVR overlay 14, second half (0x1BF8..0x3C72): choosing targets for
// spells and missiles, the attack animation and a few combat helpers.

// AimTarget is the 7-byte record the targeting routines fill in: who was
// aimed at (nil for an empty square) and the square.
type AimTarget struct {
	C    *Character
	X, Y byte
}

// SpellData returns byte k of the 16-byte spell record at DS:3200.
func (e *RuleState) SpellData(spell, k int) SpellProp { return e.T().Spells[spell&0xFF][6+k] }

// friendToHeal = ovr14:1BF8 (c, var who): the friend to heal around c: the most
// hurt of the same side on the nine squares round c (c itself only when
// under half its hit points), or the first body lying there when no one
// with fewer than 8 hit points is.
func (e *RuleState) friendToHeal(c *Character, who **Character) bool {
	var best *Character
	bestHP, body := byte(0xFF), 0
	for dir := 0; dir <= 8; dir++ {
		x := byte(e.CombatX(c)) + byte(DirDX[dir])
		y := byte(e.CombatY(c)) + byte(DirDY[dir])
		idx, ground := e.MapAt(int(x), int(y))
		if idx > 0 {
			*who = e.Combatant(byte(idx))
			t := *who
			if t != nil && t.Side == c.Side && t.HP < t.MaxHP {
				if t.HP < bestHP || (t == c && t.HP < t.MaxHP>>1) {
					best, bestHP = t, t.HP
				}
			}
		} else if ground == e.grounds().Body {
			for b := 1; b <= e.Nbodies; b++ {
				bd := e.Bodies[b]
				if bd.C != nil && (bd.C.Status < StatusDead || bd.C.Status > StatusGone) && bd.X == x && bd.Y == y {
					body = b
				}
			}
		}
	}
	if bestHP >= 8 && body != 0 {
		*who = e.Bodies[body].C
	} else {
		*who = best
	}
	return *who != nil
}

// chooseSpellTarget = ovr14:1E16 (spell, silent, allowGround, var t): one
// target for the selected character's spell. With silent 0 the player
// aims; otherwise it is picked for him. The square goes to DS:6DB9/6DBA.
func (e *RuleState) chooseSpellTarget(spell, silent, allowGround int, t *AimTarget) bool {
	sel := e.Acting()
	found := false
	if silent == 0 {
		found, *t = e.aimNow(sel, e.SpellDuration(spell), allowGround)
		sel.Combat.Target = t.C
	} else if e.SpellData(spell, 8) == 0 {
		t.C = sel
		if spell != SpCureLightWounds || e.friendToHeal(sel, &t.C) {
			t.X, t.Y = byte(e.CombatX(t.C)), byte(e.CombatY(t.C))
			found = true
		}
	} else {
		for tries := 0x14; tries > 0 && !found; tries-- {
			ok := true
			if !e.pickTarget(sel, e.SpellDuration(spell), 0, 1) {
				continue
			}
			tp := sel.Combat.Target
			if area := int(e.SpellData(spell, 9)); area > 0 {
				e.BuildNear(e.CombatX(tp), e.CombatY(tp), area, 0xFF, 1)
				for _, n := range e.Near {
					if int(e.Combatant(n.Index).Side) != B2i(IsParty(sel)) {
						ok = false
					}
				}
			}
			if e.IsHelpless(tp) {
				for i := 1; i <= 4; i++ {
					if AffectType(e.SpellData(spell, 4)) == helplessAffects[i] {
						ok = false
					}
				}
			}
			if ok {
				t.C = sel.Combat.Target
				t.X, t.Y = byte(e.CombatX(t.C)), byte(e.CombatY(t.C))
				found = true
			}
		}
	}
	if found {
		e.TgtX, e.TgtY = int(t.X), int(t.Y)
	} else {
		*t = AimTarget{}
	}
	return found
}

// chooseSpellTargets = ovr14:20C3 (spell, silent, var ok): the targets of the spell
// into DS:6C91+4i (count DS:6C94). ok is false when none was chosen.
func (e *RuleState) chooseSpellTargets(spell, silent int) (ok bool) {
	var tmp AimTarget
	ok = true
	e.NSpellTargets = 0
	e.areaSpell = 0
	e.TgtX, e.TgtY = e.CombatX(e.Acting()), e.CombatY(e.Acting())
	sel := e.Acting()
	fromNear := func(rng int) {
		e.BuildNear(int(e.TgtX), int(e.TgtY), rng, 0xFF, 1)
		for i, n := range e.Near {
			e.SetTgt(i+1, e.Combatant(n.Index))
		}
		e.NSpellTargets = len(e.Near)
		e.areaSpell = 1
	}
	kind := e.SpellData(spell, 0) & 0xF
	switch {
	case kind == 0:
		e.SetTgt(1, sel)
		e.NSpellTargets = 1
	case kind == 0xF:
		if !e.chooseSpellTarget(spell, silent, 0, &tmp) {
			ok = false
		} else if sel.Combat.Target != nil {
			e.SetTgt(1, sel.Combat.Target)
			e.NSpellTargets = 1
		} else {
			fromNear(int(e.SpellData(spell, 0)&0xF) >> 4) // always 0, as in the original
		}
	case kind >= 8 && kind <= 0xE:
		picked := e.chooseSpellTarget(spell, silent, 1, &tmp)
		if !picked {
			ok = false
		} else {
			fromNear(int(e.SpellData(spell, 0) & 7))
		}
	default:
		left := int(kind&3) + 1
		n := 0
		for left > 0 {
			if !e.chooseSpellTarget(spell, silent, 0, &tmp) {
				left--
				continue
			}
			dup := false
			for k := 1; k <= n; k++ {
				if e.tgt(k) == tmp.C {
					dup = true
				}
			}
			if !dup {
				n++
				e.SetTgt(n, tmp.C)
				left--
			} else if silent == 0 {
				e.notice("Already been targeted")
			} else {
				left--
			}
		}
		e.NSpellTargets = n
		if n == 0 {
			ok = false
		}
	}
	return
}

// CastTurnSpell = ovr14:2414 (spell, a, var quit): the selected character picks
// (spell 0: from their memorised spells) and casts a spell. Spells with a
// casting time of 3 or more only begin and are cast in a later round.
func (e *RuleState) CastTurnSpell(spell, a int, quit *bool) {
	sel := e.Acting()
	*quit = false
	if spell > 0 && e.SpellData(spell, 5) == 0 {
		e.notice("Camp Only Spell")
		spell = 0
	}
	if a == 0 {
	}
	if spell > 0 {
		delay := int(e.SpellData(spell, 6)) / 3
		if delay == 0 {
			if !e.CastSpell(SpellID(spell), byte(a), 1, CastReply{Fight: true}, quit) { // extra: a cancelled aim goes back to the menu, the turn (and its moves) go on
				*quit = e.ClearCombatInfo(sel)
			}
		} else {
			*quit = true
			e.CombatSay(sel, "Begins Casting", true)
			sel.Combat.Spell = SpellID(spell)
			if int(int8(sel.Combat.Delay)) > delay {
				sel.Combat.Delay -= byte(delay)
			} else {
				sel.Combat.Delay = 1
			}
		}
	}
}

// RangedFacing = ovr14:2574 (c, target): c may backstab the target: a thief
// with no armour or leather, no weapon or a club, dagger or broad, long or
// short sword, standing behind it (the direction from c to it is the one it faces).
func (e *RuleState) RangedFacing(c, target *Character) bool {
	flag := false
	w, armor := c.Slots[SlotWeapon], c.Slots[SlotArmor] // +0xCC, +0xD4
	var wt ItemType                                     // uninitialised in the original when there is no weapon
	if w != nil {
		wt = w.Type
	}
	if int8(c.Levels[ClsThief]) > 0 && (armor == nil || e.isKind(armor.Type, ItemIDLeatherArmor, ItemIDBracers)) { // bracers: in the armour's slot (extra, readySlot), not armour
		if w == nil || e.isKind(wt, ItemIDClub, ItemIDDagger, ItemIDBroadSword, ItemIDLongSword, ItemIDShortSword) {
			flag = true
		}
	}
	return flag && target.Combat.Attackers > 1 && byte(e.directionTo(c, target)) == target.Combat.Facing
}

// directionTo = ovr14:263D (from, to): the direction 0..7 (8 when none fits)
// from one combatant to the other.
func (e *RuleState) directionTo(from, to *Character) int {
	x1, y1 := e.CombatX(from), e.CombatY(from)
	x2, y2 := e.CombatX(to), e.CombatY(to)
	dir := 0
	for !e.InArc(x1, y1, x2, y2, dir) {
		dir++
	}
	return dir
}

// attackAnimation = ovr14:26B0 (from, to, weapon): the weapon flies from
// one combatant to the other (its sprite is present's missileArt).
func (e *RuleState) attackAnimation(from, to *Character, weapon *Item) {
	e.Sound(string(rules.CueMissile))
	e.shoot(e.CombatX(from), e.CombatY(from), e.CombatX(to), e.CombatY(to), e.kind(weapon.Type), "")
	e.Sound(string(rules.CueMissileEnd))
}

// FoesHealthPercent = ovr14:2879: DS:6E2F = the monsters' remaining hit points in
// five percent steps (20 * hp / maximum, times 5; dead ones count in the
// maximum).
func (e *RuleState) FoesHealthPercent() {
	var hp, max uint16
	for _, c := range e.Party {
		if c.Side == SideMonsters {
			if c.Okay != 0 {
				hp += uint16(c.HP)
			}
			max += uint16(c.MaxHP)
		}
	}
	if max != 0 {
		e.FoesHealth = byte(20 * hp / max * 5) // the 16-bit product, dx cleared before the divide
	}
}

// fastestFoeMove = ovr14:2908 (c): the highest moveRate / 2 over the living
// combatants on the other side.
func (e *RuleState) fastestFoeMove(c *Character) int {
	best := byte(0)
	for _, t := range e.Party {
		if B2i(IsParty(c)) == int(t.Side) && t.Okay != 0 {
			if v := byte(e.MoveRate(t)) >> 1; v > best {
				best = v
			}
		}
	}
	return int(best)
}

// mayAttack = ovr14:2998 (c, target): may c attack the target? Attacking
// a friend asks first (AsksAlly; the question is aimConfirm's); yes turns
// every sane character of the party hostile (allyAttacked).
func (e *RuleState) AsksAlly(c, target *Character) bool {
	return B2i(IsParty(target)) != int(c.Side) && c.Auto == 0
}

// AllyAttacked: mayAttack's yes.
func (e *RuleState) AllyAttacked() {
	e.AttackedAlly = 1
	e.Game.AttackedAlly = 1
	for _, m := range e.Party {
		if m.Status == StatusOkay && m.Control > 0x7F {
			m.Side = SideMonsters
			if m.Combat != nil {
				m.Combat.Target = nil
			}
		}
	}
	e.CountSides()
}

// CanTarget (aimPrompt's test, ovr14:2AA0): is Target offered for cand?
// In range, and with a missile weapon: ammunition, and no foe adjacent
// unless the weapon is thrown.
func (e *RuleState) CanTarget(c, cand *Character, rng, showRange int) bool {
	return e.TargetWhy(c, cand, rng, showRange) == ""
}

// TargetWhy (extra) is canTarget's test with the reason it fails ("" when
// it does not): in range, and with a missile weapon: ammunition, and no foe
// adjacent unless the weapon is thrown.
func (e *RuleState) TargetWhy(c, cand *Character, rng, showRange int) string {
	if e.RangeTo(c, cand) > rng {
		return outOfRange(rng)
	}
	var tmp *Item
	switch {
	case showRange == 0:
		return ""
	case cand == c:
		return "Not a target"
	case !e.RangedWeapon(c):
		return ""
	case !e.AmmoFor(c, &tmp):
		return "No ammunition"
	case e.NearEnemies(c, 1) > 0 && !e.Throwable(c):
		return "A foe is too close to shoot"
	}
	return ""
}

// outOfRange (extra): the refusal for a target farther than rng squares.
func outOfRange(rng int) string {
	if rng == 1 {
		return "Out of range: 1 square at most"
	}
	return "Out of range: " + strconv.Itoa(rng) + " squares at most"
}

// AimCheck (extra) is aimPrompt's and aimManual's test of the square
// (x, y) as the aim: who would be the target, and why it cannot be taken
// ("" when it can). A combatant in sight (list) is tested as aimPrompt
// does, any other square as aimManual's cursor.
func (e *RuleState) AimCheck(c *Character, list []NearEntry, x, y, rng, showRange, allowGround int) (*Character, string) {
	who, ground := e.MapAt(x, y)
	if who > 0 && slices.ContainsFunc(list, func(n NearEntry) bool { return n.Index == byte(who) }) {
		cand := e.Combatant(byte(who))
		if why := e.TargetWhy(c, cand, rng, showRange); why != "" {
			return cand, why
		}
		if !e.MayActAgainst(c, cand) {
			return cand, "Cannot act against " + cand.Name()
		}
		return cand, ""
	}
	dist, tx, ty := 0xFF, x, y
	if !e.LineOfSight(&dist, &tx, &ty, e.CombatX(c), e.CombatY(c)) {
		return nil, "Out of sight"
	}
	dist >>= 1
	var cand *Character
	if who > 0 {
		cand = e.Combatant(byte(who))
	} else if ground == e.grounds().Body {
		for b := 1; b <= e.Nbodies; b++ {
			if int(e.Bodies[b].X) == x && int(e.Bodies[b].Y) == y {
				cand = e.Bodies[b].C
			}
		}
	}
	switch {
	case rng < dist:
		return cand, outOfRange(rng)
	case e.GroundInfo(ground)[0] == 0xFF:
		return cand, "A wall is in the way"
	case cand == nil && allowGround == 0:
		return nil, "No one there to target"
	case cand == nil:
		return nil, ""
	case !e.MayActAgainst(c, cand):
		return cand, "Cannot act against " + cand.Name()
	case showRange != 1:
		return cand, ""
	case cand == c || who == 0 && ground == e.grounds().Body:
		return cand, "Not a target"
	}
	if e.RangedWeapon(c) {
		var tmp *Item
		if !e.AmmoFor(c, &tmp) {
			return cand, "No ammunition"
		} else if e.NearEnemies(c, 1) > 0 && !e.Throwable(c) {
			return cand, "A foe is too close to shoot"
		}
	}
	return cand, ""
}

// pickTarget = ovr14:37E4 (c, rng, a, b): a target for c's attack from the
// enemies within rng, kept in c's combat block. Up to two passes of 20
// random picks; the second sees through walls (not with b set).
func (e *RuleState) pickTarget(c *Character, rng, a, b int) bool {
	first, found, second := false, false, false
	tgt := c.Combat.Target
	if b != 0 || tgt == nil || tgt.Side == c.Side || tgt.Okay == 0 || !e.MayActAgainst(c, tgt) {
		c.Combat.Target = nil
	}
	if c.Combat.Target != nil {
		found = true
	}
	for !found && !second {
		second = first
		if first && b == 0 {
			e.Cmap.seeAll = 1
		}
		tries := 0x14
		n := e.NearEnemies(c, rng)
		for tries > 0 && !found && n > 0 {
			tries--
			r := e.Dice(1, n)
			if e.Targets[r] > 0 {
				cand := e.Combatant(byte(e.Targets[r]))
				if (a != 0 && e.Cmap.seeAll != 0) || e.MayActAgainst(c, cand) {
					found = true
					c.Combat.Target = cand
				} else {
					e.Targets[r] = 0
					for i := 1; i <= n; i++ {
						if e.Targets[i] > 0 {
							found = true // sic, the original sets the result flag
						}
					}
					if found {
						found = false
					} else {
						n = 0
					}
				}
			}
		}
		if !first {
			first = true
		}
	}
	e.Cmap.seeAll = 0
	return found
}

// rake = ovr14:39D2 (c, affect, flag): the affect handler of an
// attacker's rake: while DS:6823 is 2 and its victim is up, c rakes it.
func (e *RuleState) rake(c *Character, a *Affect, flag byte) {
	if e.slotHits[1] == 2 && c.Combat.Target != nil && c.Combat.Target.Okay != 0 {
		c.Combat.AttackSlot = 2
		c.Attacks[1], c.Dice[1], c.Sides[1], c.Attacks[0] = 2, 2, 4, 0
		e.CombatSay(c, "Rakes...", true)
		e.FullAttack(c, c.Combat.Target, 0, nil)
		e.ClearCombatInfo(c)
	}
}

// strikeTied = ovr14:3A86 (c, affect, flag): the affect handler tying c to
// the combatant a.Data (DS:6826): c strikes at it until one of them is
// down or the affect is removed.
func (e *RuleState) strikeTied(c *Character, a *Affect, flag byte) {
	e.tiedTarget = e.Combatant(a.Data)
	if flag != 0 || c.Okay == 0 || e.tiedTarget.Okay == 0 {
		e.RemoveAffect(e.tiedTarget, affHeldFast, nil)
		if flag == 0 {
			a.Call = 0
			e.RemoveAffect(c, affStrikeTied, a)
		}
		return
	}
	c.Attacks[0], c.Attacks[1], c.Dice[0], c.Sides[0] = 1, 0, 3, 4
	e.FullAttack(c, e.tiedTarget, 1, nil)
	e.ClearCombatInfo(c)
	if e.tiedTarget.Okay == 0 {
		e.RemoveAffect(c, affStrikeTied, a)
	}
}
