package rules1e

// Combat turn, first half (ovr14:0000..1BF8): setting up a combatant's turn,
// movement, attacks of opportunity, fleeing, sweeps, turning undead and the
// attack itself. The second half (from 1BF8) is ported separately.

// ds2 reads the word at DS:ofs of the exe image.

import (
	"gobox/core"
	"gobox/core/dnd"
	"gobox/core/goldbox"
	"strconv"
)

// TurnSetup = ovr14:0000 (char): fill the combat block for a new round.
func (e *RuleState) TurnSetup(c *Character) {
	cb := c.Combat
	cb.Spell, cb.CanCast, cb.CanUse, cb.Attacked, cb.AttackSlot = 0, 1, 1, 0, 2
	e.SetSpeed(c)
	e.turnValue = byte(c.BaseAttack[1])
	e.modMoves = 0
	e.AffectGroup(c, 0x12)
	c.Attacks[1] = Count(AttacksFromRate(int(e.turnValue), int(e.CombatRound)))
	cb.Sweeps = c.FightLevel
	if c.Okay != 0 {
		cb.Delay = byte(int(goldbox.DexMissile(byte(c.Stats[StatDex]))) + e.Dice(1, 6))
		if int8(cb.Delay) < 1 {
			cb.Delay = 1
		}
		if byte((int(int8(c.Side))+1)&int(e.Game.Surprise)) != 0 { // only the low byte is tested
			cb.Delay -= 6
		}
		if int8(cb.Delay) < 0 || int8(cb.Delay) > 0x14 {
			cb.Delay = 0
		}
	} else {
		cb.Delay = 0
	}
	cb.Moves = byte(e.MoveRate(c))
	e.Rules().RoundStart(e, c)
}

// MoveRate = ovr14:0123 (char): movement rate of the turn (half of it is kept in +0x113
// by setSpeed); the affects of group 0x12 may change it.
func (e *RuleState) MoveRate(c *Character) int {
	v := c.Move
	if c.Side == SideParty {
		v += byte(e.Game.PartyMove)
	}
	e.modMoves = 1
	if v < 1 || v > 0x60 {
		v = 1
	}
	e.turnValue = v << 1
	e.AffectGroup(c, 0x12)
	e.modMoves = 0
	return int(e.turnValue)
}

// DamageRoll = ovr14:0191 (att, vic, slot): damage roll of the attack slot (1 or 2)
// (the rule set's Dealt), tripled-or-so for a backstab, then the
// affects of groups 4 (attacker) and 5 (victim) may change DS:6818.
func (e *RuleState) DamageRoll(a, b *Character, slot int) {
	h := Hit{Dice: int(a.Dice[slot-1]), Sides: int(a.Sides[slot-1]), Bonus: int(int8(a.Bonus[slot-1]))}
	h.Roll = e.RollDice(h.Dice, h.Sides)
	if slot == 1 { // weaponStats: the strength adds to the readied weapon's hand
		_, h.Ability, _ = e.WeaponAbility(a)
	}
	// extra: the rule set's; the original holds it in a byte and compares it
	// with 0 "above or equal": a negative roll wraps (kept, T4d)
	e.Rv.Damage = int16(byte(e.Rules().Dealt(e, a, b, h)))
	if e.RangedFacing(a, b) && e.Rules().Backstabs(e) {
		// thief level div 4 + 2, as the original (Felix's ruling 17: the higher multiple)
		e.Rv.Damage = int16(int(e.Rv.Damage) * (int(a.Levels[ClsThief])>>2 + 2))
	}
	e.Rv.DmgType = 0
	e.AffectGroup(a, 4)
	e.AffectGroup(b, 5)
}

// attackText = ovr14:02F9 (att, vic, kind, damage, shown, hit): the attack text. kind 1 is
// a plain hit, 2 a backstab, 3 a kill of a helpless target; damage is
// applied, shown printed; hit 0 is a miss.
func (e *RuleState) attackText(a, b *Character, kind, damage, shown, hit int) {
	var s string
	switch kind {
	case 2:
		s = "-Backstabs-"
	case 3:
		s = "slays helpless"
	default:
		s = "Attacks"
	}
	e.CombatSay(a, s, true)
	e.factAttack(a, b, hit != 0, shown&0xFF)
	s = ""
	if kind == 1 {
		s = "(from behind) "
	}
	if hit == 0 {
		s += "and Misses"
	} else {
		if kind == 3 {
			s = "with one cruel blow"
		} else {
			s += "Hitting for " + strconv.Itoa(shown&0xFF)
			if shown&0xFF == 1 {
				s += " point "
			} else {
				s += " points "
			}
			s += "of damage"
		}
		e.ApplyDamage(b, damage&0xFF)
	}
	e.CombatSay(b, s, false)
	if damage&0xFF > 0 {
		b.Combat.CanCast = 0
		if b.Combat.Spell > 0 {
			e.CombatSay(b, "lost a spell", true)
			e.forgetSpell(b, SpellID(b.Combat.Spell))
			b.Combat.Spell = 0
		}
	}
	if b.Okay == 0 {
		e.CombatSay(b, "goes down", true)
		if b.Status == StatusDying {
			e.CombatSay(nil, "and is Dying", false)
		}
		if st := b.Status; st >= StatusDead && st <= StatusGone { // set at CS:02CF
			e.CombatSay(b, "is killed", true)
		}
		e.removeCombatAffects(b)
		e.AffectGroup(b, 0xD)
		if b.Okay == 0 {
			e.combatRemoveFromMap(b)
		}
	}
}

// waitingFoesStrike = ovr14:0627 (char): the enemies next to a mover that were waiting (+7) lose
// their reaction, turn to it and attack it.
func (e *RuleState) waitingFoesStrike(c *Character) {
	n := e.NearEnemies(c, 1)
	for i := 1; i <= n; i++ {
		if c.Okay == 0 {
			continue
		}
		t := e.Combatant(byte(e.Targets[i]))
		if t.Combat.Guarding == 0 || e.IsHelpless(t) {
			continue
		}
		t.Combat.Guarding = 0
		e.CountAttackTurn(t, c)
		e.FullAttack(t, c, 0, nil)
	}
}

// stepSquare = ovr14:0712 (dir, char): step one square (2 movement points, 3 diagonally).
func (e *RuleState) stepSquare(c *Character, dir int) {
	cb := c.Combat
	cost := byte(2)
	if dir&1 != 0 {
		cost = 3
	}
	if e.Fight.forced { // extra: MoveTo's: a shift or forced movement spends no moves
		cost = 0
	}
	if cost > cb.Moves {
		cb.Moves = 0
		return
	}
	cb.Moves -= cost
	if !e.Fight.forced {
		e.Rules().Moved(e, c) // extra: the rule set's
	}
	idx := e.CombatIndex(c)
	ox, oy := e.Cpos[idx].X, e.Cpos[idx].Y
	nx := ox + byte(DirDX[dir])
	ny := oy + byte(DirDY[dir])
	e.Cpos[idx].X, e.Cpos[idx].Y = nx, ny
	if e.Fight.On {
		e.Emit(core.Walked{Unit: unit(idx), Path: []core.BoardPos{{X: int(nx), Y: int(ny)}}}) // extra: each step held, so a walk (the computer's too: Quick, monsters) plays square by square
	}
	e.CombatRedrawMap()
	cb.Attackers, cb.Flank = 0, 0
	e.Sound(string(core.CueStep))
	e.Rules().Entered(e, c) // extra: the rule set's (a zone)
	e.waitingFoesStrike(c)
	if c.Okay == 0 || e.IsHelpless(c) {
		cb.Moves = 0
	}
}

// freeAttacksOnLeave = ovr14:08B2 (dir, char): the mover steps (virtually) one square in dir; the
// enemies that were next to it and no longer are get a free attack.
func (e *RuleState) freeAttacksOnLeave(c *Character, dir int) {
	var lst [0x0D]Slot // bp-0xD+i, i = 1..: the DS:6DE4 list before the move
	idx := e.CombatIndex(c)
	n := e.NearEnemies(c, 1)
	if n == 0 {
		return
	}
	for i := 1; i <= n; i++ {
		lst[i] = e.Targets[i]
	}
	dx, dy := byte(DirDX[dir]), byte(DirDY[dir])
	e.Cpos[idx].X += dx
	e.Cpos[idx].Y += dy
	m := e.NearEnemies(c, 1)
	e.Cpos[idx].X -= dx
	e.Cpos[idx].Y -= dy
	for i := 1; i <= n; i++ {
		found := false
		for j := 1; j <= m; j++ {
			if e.Targets[j] == lst[i] {
				found = true
			}
		}
		if found {
			lst[i] = 0
		}
	}
	for i := 1; i <= n; i++ {
		if lst[i] == 0 || c.Okay == 0 {
			continue
		}
		t := e.Combatant(byte(lst[i]))
		if e.IsHelpless(t) || !e.MayActAgainst(t, c) || !e.Rules().Provokes(e, c, t) { // extra: the rule set's (a ranger's Hit and Run)
			continue
		}
		if e.FindAffect(t, affBloodDrain) != nil || e.FindAffect(t, affStrikeTied) != nil {
			continue
		}
		attacked := false
		tcb := t.Combat
		for k := int(tcb.Facing) + 6; k <= int(tcb.Facing)+10; k++ {
			if attacked {
				continue
			}
			if int8(tcb.Delay) <= 0 && tcb.Attackers != 0 {
				if !e.InArc(e.CombatX(t), e.CombatY(t), e.CombatX(c), e.CombatY(c), k%8) {
					continue
				}
			}
			w := 1
			if t.BaseAttack[0] == 0 {
				w = 2
			}
			for s := 1; s <= 2; s++ {
				if t.Attacks[s-1] > 0 {
					w = s
				}
			}
			if t.Attacks[w-1] == 0 {
				t.Attacks[w-1] = 1
			}
			tcb.AttackSlot = byte(w)
			saved := t.Combat.Target
			e.Fight.free = true // extra: Opportunity
			e.FullAttack(t, c, 1, nil)
			e.Fight.free = false
			attacked = true
			t.Combat.Target = saved
			if c.Okay != 0 {
			}
		}
	}
}

// TryFlee = ovr14:0C6A (char) bool: try to run away; "Got Away" removes the character.
func (e *RuleState) TryFlee(c *Character) bool {
	away := false
	if e.NearEnemies(c, 0xFF) == 0 {
		away = true
	} else {
		a := byte(e.MoveRate(c)) >> 1
		b := byte(e.fastestFoeMove(c))
		if b < a {
			away = true
		} else if b == a && e.Dice(1, 2) == 1 {
			away = true
		}
	}
	if away {
		e.KillCharacter(c, StatusRunning, "Got Away")
	} else {
		e.notice("Escape is blocked")
	}
	return e.ClearCombatInfo(c)
}

// SetSpeed = ovr14:0D23 (char): the character's speed in +0x113 from its movement base
// (+0xA1) or the readied weapon's speed (DS:5582 record +5, at least 2).
func (e *RuleState) SetSpeed(c *Character) {
	flag := false
	old := c.Attacks[0]
	c.Attacks[0] = c.BaseAttack[0]
	var it *Item
	if e.RangedWeapon(c) && e.AmmoFor(c, &it) {
		flag = true
		v := byte(0) // (the original reads the item at +0xCC unchecked)
		if c.Slots[SlotWeapon] != nil {
			v = e.Base(c.Slots[SlotWeapon].Type).Speed
		}
		if v < 2 {
			v = 2
		}
		e.turnValue = v
	} else {
		e.turnValue = byte(c.Attacks[0])
	}
	e.modMoves = 0
	e.AffectGroup(c, 0x12)
	r := byte(AttacksFromRate(int(e.turnValue), int(e.CombatRound)))
	if flag && it != nil {
		w := byte(1)
		if it.Count > w {
			w = it.Count
		}
		if w < r {
			r = w
		}
	}
	// fixed, as the Pascal reference: once the turn has attacked, the original returned here with
	// +0x113 left at the full base rate (or raised it to r), so View or Use
	// after a kill handed out attacks again; the attacks left never grow
	if c.Combat.Attacked != 0 {
		r = min(r, byte(old))
	}
	c.Attacks[0] = Count(r)
}

// Sweep = ovr14:0E82 (att, vic) bool: a fighter with more attacks than... sweeps: the
// attacker (fewer attacks than +5 allows, victim not a "big" target) hits
// several adjacent weak enemies, one attack each.
func (e *RuleState) Sweep(a, b *Character) bool {
	res := false
	acb := a.Combat
	if byte(a.Attacks[0]) < acb.Sweeps && b.MaxLevel == 0 && e.RangeTo(a, b) == 1 {
		n := e.NearEnemies(a, 1)
		cnt := 0
		bi := 0
		for i := 1; i <= n; i++ {
			t := e.Combatant(byte(e.Targets[i]))
			if t == b {
				bi = i
			}
			if t.MaxLevel == 0 {
				cnt++
			}
		}
		if Count(cnt) > a.Attacks[0] {
			if byte(cnt) > acb.Sweeps {
				cnt = int(acb.Sweeps)
			}
			e.CombatSay(a, "sweeps", true)
			if e.Combatant(byte(e.Targets[1])) != b {
				e.Targets[bi] = e.Targets[1] // (bi 0 when b was not found: DS:6DE4)
				e.Targets[1] = Slot(e.CombatIndex(b))
			}
			for i := 1; i <= n; i++ {
				if cnt > 0 {
					t := e.Combatant(byte(e.Targets[i]))
					if t.MaxLevel == 0 {
						e.CountAttackTurn(a, t)
						a.Attacks[0] = 1
						e.FullAttack(a, t, 0, nil)
						cnt--
					}
				}
			}
			res = true
		}
	}
	return res
}

// MayActAgainst = ovr14:1085 (t, c) bool: may t act against c? Affect group 1 of c can veto
// (DS:681E); then group 0 of t is run with c as its target.
func (e *RuleState) MayActAgainst(t, c *Character) bool {
	if c == nil {
		return false
	}
	if t == c {
		return true
	}
	e.AffectVeto = 0
	e.AffectGroup(c, 1)
	if e.AffectVeto == 0 {
		saved := t.Combat.Target
		t.Combat.Target = c
		e.AffectGroup(t, 0)
		t.Combat.Target = saved
	}
	return e.AffectVeto == 0
}

// ClericTurnUndead = ovr14:116A (char): a cleric turns undead (ovr14:116A).
func (e *RuleState) ClericTurnUndead(c *Character) {
	e.CombatSay(c, "turns undead...", true)
	any, stop := false, false
	c.Combat.TurnUsed = 1
	limit := 6
	hd := e.Dice(1, 0xC)
	roll := e.Dice(1, 0x14)
	var cls int
	cls = TurnColumn(int(int8(c.Levels[ClsCleric])))
	var t *Character
	for e.weakestUnturned(c, &t) && hd > 0 && !stop {
		x := e.T().TurnNeed(int(t.Undead), cls)
		ax := x
		if ax < 0 {
			ax = -ax
		}
		if roll < ax {
			stop = true
			continue
		}
		any = true
		if x > 0 {
			t.Combat.Turned = 1
			e.hitMessage(t, true, "is turned")
		} else {
			e.CombatSay(t, "Is destroyed", true)
			e.combatRemoveFromMap(t)
			t.Status = StatusGone
			t.Okay = 0
		}
		if limit > 0 {
			limit--
		}
		hd--
		if hd == 0 && limit > 0 && x < 0 {
			hd++
		}
	}
	if !any {
		e.notice("Nothing Happens...")
	}
	e.CountSides()
	e.ClearCombatInfo(c)
}

// weakestUnturned = ovr14:1351 (var out, char) bool: the weakest (lowest +0x76 > 0) enemy not
// yet turned, within reach.
func (e *RuleState) weakestUnturned(c *Character, out **Character) bool {
	best := int8(0xD)
	*out = nil
	n := e.NearEnemies(c, 0xFF)
	found := false
	for i := 1; i <= n; i++ {
		t := e.Combatant(byte(e.Targets[i]))
		if t.Combat.Turned == 0 && int8(t.Undead) > 0 && int8(t.Undead) < best {
			best = int8(t.Undead)
			*out = t
			found = true
		}
	}
	return found
}

// attackSlots = ovr14:1406 (att, vic, w, var done): the attacker's attack slots against the
// victim; done is 1 when the attacker has used all its attacks.
func (e *RuleState) attackSlots(a, b *Character, w int) (done bool) {
	acb := a.Combat
	done = false
	e.slotHits[1], e.slotHits[2] = 0, 0
	e.SlotShots[1], e.SlotShots[2] = 0, 0
	any, stop := false, false
	e.Rv.Damage = 0
	acb.Attacked = 1
	if e.Rules().Attacking(e, a, b); a.Okay == 0 || e.IsHelpless(a) { // extra: the rule set's interrupt struck it down
		e.Fight.Armed = Armed{}
		return e.ClearCombatInfo(a)
	}
	if e.IsHelpless(b) {
		e.Sound("hit_superb")
		for acb.AttackSlot > 0 && a.Attacks[acb.AttackSlot-1] == 0 {
			acb.AttackSlot--
		}
		e.SlotShots[acb.AttackSlot]++
		e.attackText(a, b, 3, (int(b.HP)+5)&0xFF, 1, 1)
		e.removeInvisibility(a)
		for s := 1; s <= 2; s++ {
			a.Attacks[s-1] = 0
		}
		any = true
		done = true
	} else {
		if a.Slots[SlotWeapon] != nil && (b.Size > 0x80 || b.Size&7 > 1) {
			typ := a.Slots[SlotWeapon].Type
			a.Dice[0] = Count(e.Base(typ).LargeDice)
			a.Sides[0] = Count(e.Base(typ).LargeSides)
			a.Bonus[0] = Modifier(int(int8(a.Bonus[0])) - int(int8(e.Base(typ).Bonus)) + int(int8(e.Base(typ).LargeBonus)))
		}
		e.ApplyItems(b)
		e.AffectGroup(b, 0xB)
		hand := 0
		var need byte
		if e.RangedFacing(a, b) {
			need = byte(b.AC[1] - 2)
		} else {
			if b.Combat.Attackers > 1 && byte(e.directionTo(a, b)) == b.Combat.Facing && b.Combat.Flank > 4 {
				w = 1
			}
			if w != 0 {
				hand = 1
			}
			need = byte(b.AC[hand])
		}
		bonus := int(need)
		e.RangePenalty(a, b, &bonus)
		need = byte(e.Rules().HitNeed(e, a, b, bonus))
		mode := 0
		if w != 0 {
			mode = 1
		}
		if e.RangedFacing(a, b) {
			mode = 2
		}
		for s := int(acb.AttackSlot); s >= 1; s-- {
			for a.Attacks[s-1] != 0 && !stop {
				a.Attacks[s-1]--
				acb.AttackSlot = byte(s)
				e.SlotShots[s]++
				if e.AttackHits(a, b, int(need)) || e.IsHelpless(b) {
					e.slotHits[s]++
					e.Sound(e.hitEvent(b, "hit")) // extra: the weapon-hit sound (the original's bump is barely audible)
					any = true
					e.DamageRoll(a, b, s)
					e.attackText(a, b, mode, int(e.Rv.Damage), int(e.Rv.Damage), 1)
					if b.Okay != 0 && e.Rules().Special(e, a) {
						e.AffectGroup(a, s+1)
					}
					if b.Okay == 0 {
						stop = true
					}
				}
			}
		}
		if !any {
			e.Sound(string(core.CueMiss))
			e.attackText(a, b, mode, 0, 0, 0)
		}
		done = true
		for s := 1; s <= 2; s++ {
			if a.Attacks[s-1] > 0 {
				done = false
			}
		}
		acb.Sweeps = 0
	}
	e.Rules().AfterAttack(e, a, b, any)
	if e.Fight.Armed.C == a { // a maneuver rides one attack
		e.Fight.Armed = Armed{}
	}
	if done {
		done = e.ClearCombatInfo(a)
	}
	return
}

// CountAttackTurn = ovr14:17F6 (att, vic): the victim counts one more attack and turns towards
// the attacker (the turned amount accumulates in +0x12).
func (e *RuleState) CountAttackTurn(att, vic *Character) {
	cb := vic.Combat
	cb.Attackers++
	d := (e.directionTo(vic, att) - int(cb.Facing) + 8) % 8
	if d > 4 {
		d = 8 - d
	}
	cb.Flank = byte((int(cb.Flank) + d) % 8)
}

// FullAttack = ovr14:1884 (att, vic, flag, item, var done): a full attack of att on vic
// (turning both, then the slots of attackSlots). item is a thrown item that is
// spent; done is set when the attacker's turn ends.
func (e *RuleState) FullAttack(a, b *Character, flag int, it *Item) (done bool) {
	if b.Combat.Attackers < 3 && flag == 0 {
		e.TurnTo(b, byte(e.directionTo(b, a)))
	}
	dir := byte(e.directionTo(a, b))
	e.TurnTo(a, dir) // the attack's pose is the presenter's (Attacked)
	a.Combat.Target = b
	if it != nil {
		e.attackAnimation(a, b, it)
	}
	if a.Slots[SlotWeapon] != nil && e.isKind(a.Slots[SlotWeapon].Type, dnd.ItemSling) {
		e.attackAnimation(a, b, a.Slots[SlotWeapon])
	}
	done = true
	if a.Attacks[0] > 0 || a.Attacks[1] > 0 {
		saved := e.Who
		for i, c := range e.Party {
			if c == a {
				e.Who = i
			}
		}
		done = e.attackSlots(a, b, flag)
		if it != nil {
			if it.Count > 0 {
				it.Count -= byte(e.SlotShots[1])
			}
			if it.Count == 0 {
				if e.Throwable(a) && it.Spells[2] != ItemPower(affEndsSpiritualHammer) {
					cp := *it
					n := &cp
					e.Treasure = append([]*Item{n}, e.Treasure...)
					n.Readied = 0
					e.RemoveItem(a, it)
					e.ThrownCopy = n
				} else {
					e.RemoveItem(a, it)
				}
			}
		}
		e.ApplyItems(a)
		e.Who = saved
	}
	if done {
		done = e.ClearCombatInfo(a)
	}
	return
}

// RangePenalty = ovr14:1B1E (att, vic, var need): the attacker's range band: every full
// weapon range band beyond the first adds 2, then 3, to the number to hit.
func (e *RuleState) RangePenalty(a, b *Character, need *int) {
	d := e.RangeTo(a, b)
	f := d
	if e.RangedWeapon(a) {
		f = (int(e.Base(a.Slots[SlotWeapon].Type).Range) - 1) / 3
	}
	if d-f > 0 {
		d -= f
		*need += 2
	}
	if d-f > 0 {
		d -= f
		*need += 3
	}
}
