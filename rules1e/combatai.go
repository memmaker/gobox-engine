package rules1e

import (
	"gobox/rules"
	"slices"
)

// The computer-controlled turn, ported from ovr10: monster AI, and party
// members under AI control or fleeing. Called from the combat loop (ovr09).

// ComputerTurn is one computer turn and what it showed, in order: the
// engine's Decide then Resolve, compared at whole turns (TestDecideAgainstPascal).
// key is the key the player pressed before it, which the turn's first poll
// takes (aiKeys).
func (e *RuleState) ComputerTurn(c *Character, key AiKey) rules.Outcome {
	n := len(e.Fx)
	e.Fight.Key = key
	e.aiTurn(c)
	e.Fight.Key = AiKeyNone
	out := rules.Outcome{Effects: slices.Clone(e.Fx[n:])}
	e.Fx = e.Fx[:n]
	return out
}

// aiTurn = ovr10:000F (char): one computer turn. Picks the facing of
// the attack, then spells, then moves and attacks until the turn is used.
func (e *RuleState) aiTurn(c *Character) {
	done := e.AiKeys(c)
	if c.Okay == 0 {
		done = e.ClearCombatInfo(c)
	}
	ci := c.Combat
	d := ci.Pattern
	if d == 0 || d > 4 || e.Dice(1, 4) == 1 {
		d = byte(e.Dice(1, 8))
		if d == 8 {
			d = byte(e.Dice(1, 2) + 4)
		} else {
			d = byte(e.Dice(1, 4)) // the first roll is thrown away
		}
	}
	ci.Pattern = d
	if !done {
		done = e.aiFleeCheck(c)
	}
	if ci.Fleeing != 0 && ci.Turned == 0 {
		e.Emit(rules.Sounded{Cue: "flee"})
		e.CombatSay(c, "flees in panic", true)
	}
	if done {
		return
	}
	// As the original (and the Pascal reference): an item first, then a spell
	// begun last turn. The computer never aims a spell where it is caught
	// itself (aiRadius rejects such targets), so its readied spell cannot be
	// lost twice by its own damage.
	if e.aiCastFromItem(c) {
		e.ClearCombatInfo(c)
		return
	}
	if ci.Spell > 0 {
		e.CastSpell(ci.Spell, 1, 1, CastReply{Fight: true}, &done)
		done = e.ClearCombatInfo(c)
		return
	}
	if e.aiTryCast(c) {
		e.ClearCombatInfo(c)
		return
	}
	if e.aiCastKnown(c) {
		return
	}
	e.AiPickWeapon(c)
	done = e.AiKeys(c)
	for !done {
		if e.pickTarget(c, 0xFF, 1, 0) && int8(ci.Delay) > 0 {
			done = e.AiAttackLoop(c)
		} else {
			done = e.aiGiveUp(c)
		}
	}
}

// aiTryCast = ovr10:0203 (char): a spell caster with a spell picked
// (weakestUnturned) casts it.
func (e *RuleState) aiTryCast(c *Character) bool {
	r := false // uninitialised in the original when the test fails: 0
	if c.Combat.TurnUsed == 0 && int8(c.Levels[ClsCleric]) > 0 {
		var t *Character
		if e.weakestUnturned(c, &t) {
			r = true
			e.ClericTurnUndead(c)
		}
	}
	return r
}

// canCast = ovr10:02EB (char, spell, level): may the computer cast the
// spell at this level: the level fits, and an area spell (DS:3209 set)
// must not catch anyone of the wrong side.
func (e *RuleState) canCast(c *Character, spell, level int) bool {
	tbl := e.T().Spells[spell][13:16]
	if int(tbl[0]) < level {
		return false
	}
	if spell != SpCureLightWounds && tbl[1] == 0 {
		return true
	}
	if spell == SpCureLightWounds {
		var t *Character
		if e.friendToHeal(c, &t) {
			return true
		}
	}
	n := e.NearEnemies(c, e.SpellDuration(spell))
	if n == 0 {
		return false
	}
	if tbl[2] == 0 {
		return true
	}
	// ovr10:0255 (nested): is anyone near (x,y) not on the selected side
	wrongSide := func(x, y int) bool {
		r := false
		e.BuildNear(x, y, int(e.T().Spells[spell][15]), 0xFF, 1)
		sel := B2i(IsParty(e.Acting()))
		for i := 1; i <= len(e.Near); i++ {
			if int(e.Combatant(e.Near[i-1].Index).Side) != sel {
				r = true
			}
		}
		return r
	}
	for i := 1; i <= n; i++ {
		p := e.Cpos[e.Targets[i]]
		if wrongSide(int(p.X), int(p.Y)) {
			return false
		}
	}
	return true
}

// aiCastFromItem = ovr10:03E4 (char): try to use an item that casts a
// spell the computer can cast; true when one was used.
func (e *RuleState) aiCastFromItem(c *Character) bool {
	r := false
	var found *Item
	lvl := 7
	rounds := e.Dice(1, 7)
	if c.Combat.CanUse == 0 || e.Rv.Side[B2i(IsParty(c))] == 0 || e.Area.NoMagic != 0 {
		return false
	}
	for k := 1; k <= rounds; k++ {
		for _, p := range c.Items {
			if found != nil {
				break
			}
			sp := p.Spells[1]
			if castsSpell(e, p) && p.Readied != 0 {
				if e.canCast(c, int(sp), lvl) {
					found = p
				}
			}
		}
		lvl--
	}
	if found != nil {
		e.UsingItem = true // useThen without its questions: the computer's item is no scroll
		e.UseItem(found, SpellID(found.Spells[1]), CastReply{Fight: true})
		r = true
	}
	return r
}

// aiCastKnown = ovr10:0545 (char): pick a known spell (c+0x17..) the
// computer may cast, trying lower levels each round; true when the spell
// call (castSpell) returned a non-empty string.
func (e *RuleState) aiCastKnown(c *Character) bool {
	var arr [MemorySlots]SpellID // the original has 20 bytes: the 21st write lands on a spare local
	n := 0
	if c.Combat.CanCast != 0 {
		for i := 0; i <= 0x14; i++ {
			if c.Memorized[i].Held() { // memorized, not still to be learned
				arr[n] = c.Memorized[i]
				n++
			}
		}
	}
	found := 0
	lvl := 7
	rounds := e.Dice(1, 7)
	if n > 0 && (c.Control > 0x7F || e.MagicOn != 0) && e.Rv.Side[B2i(IsParty(c))] != 0 {
		for it := 1; it <= rounds && found == 0; it++ {
			for t := 1; t < 4 && found == 0; t++ {
				sp := int(arr[e.Dice(1, n)-1])
				if e.canCast(c, sp, lvl) {
					found = sp
				}
			}
			lvl--
		}
	}
	if found > 0 {
		var ok bool
		e.CastTurnSpell(found, 1, &ok)
		return ok
	}
	return false
}

// aiStep: step i (1..5) of a step pattern. i 6 means no step was found;
// the original reads past the table there (DS:02D2) and the value is unused.
func aiStep(pattern byte, i int) Dir {
	if i < 1 || i > 5 {
		return 0
	}
	return aiStepPatterns[pattern][i-1]
}

// aiCanStep = ovr10:0675 (char, i, dir, var blocked): can the combatant
// step in direction DS:02AE[facing*5+i]+dir; blocked is set when the
// square is off the map.
func (e *RuleState) aiCanStep(c *Character, i, dir int) (ok, blocked bool) {
	ci := c.Combat
	r := false
	t := int(aiStep(ci.Pattern, i))
	d := (t + dir) % 8
	who, ground := e.GroundAt(c, d)
	if ground == 0 {
		return false, true
	}
	if who != 0 {
		return false, false
	}
	gc := e.T().Ground[ground][0]
	if byte(gc) >= ci.Moves {
		return false, false
	}
	cost := 2
	if d&1 != 0 {
		cost = 3
	}
	cost = int(byte(cost * int(gc)))
	if ground == e.grounds().StinkingCloud {
		if e.FindAffect(c, affAnimated) == nil && e.FindAffect(c, affCoughing) == nil &&
			e.FindAffect(c, affBlockPoisonHold) == nil && e.FindAffect(c, affBlockMind) == nil {
			cost = int(byte(ci.Moves + 1))
		}
	}
	if int(ci.Moves) >= cost {
		r = true
	}
	return r, false
}

// aiMoveStep = ovr10:07F2 (char): show the movement left and take one
// step towards the target (or a wandering direction when fleeing).
func (e *RuleState) aiMoveStep(c *Character) {
	ci := c.Combat
	if e.AiKeys(c) {
		return
	}
	if ci.Moves>>1 <= 0 || int8(ci.Delay) <= 0 || (ci.Fleeing == 0 && c.Class == ClsMage) {
		e.aiGiveUp(c)
		return
	}
	var dir int
	if ci.Fleeing == 0 {
		dir = e.directionTo(c, ci.Target)
		if e.Settings.PathSearch { // option (ruling 3, default off): the walk by path search instead of the original's step
			if pd := e.AiPathDir(c, ci.Target); pd >= 0 {
				dir, ci.Pattern = pd, 1 // pattern 1 tries straight on first
			}
		}
	} else {
		ci.Pattern = byte(e.Dice(1, 2))
		f := int(e.Facing)
		dir = f - ((f+2)%4)>>1
		if c.Side == SideParty {
			dir += 4
		}
		dir %= 8
		dir = int(byte(dir))
	}
	var ok, blocked bool
	done := false
	i := 1
	for i < 6 && !done {
		if ok, blocked = e.aiCanStep(c, i, dir); ok {
			break
		}
		if ci.Fleeing != 0 && blocked {
			done = e.TryFlee(c)
		} else {
			i++
		}
	}
	var d int
	if done {
		ci.Moves = 0
		ci.Fleeing = 0
		done = e.ClearCombatInfo(c)
	} else {
		d = (int(aiStep(ci.Pattern, i)) + dir) % 8
		if i == 6 || int(e.aiLastDir) == (d+4)%8 {
			e.aiBlocked++
			ci.Pattern = byte(int(ci.Pattern)%6 + 1)
			if e.aiBlocked > 1 {
				ci.Target = nil
				if e.aiBlocked > 2 {
					ci.Moves = 0
					done = true
				} else if !e.pickTarget(c, 0xFF, 1, 0) {
					done = e.aiGiveUp(c)
				}
			}
		}
		if i < 6 {
			e.aiLastDir = byte(d)
		} else {
			done = true
		}
	}
	if done {
		return
	}
	e.TurnTo(c, byte(d))
	e.freeAttacksOnLeave(c, int(ci.Facing))
	if c.Okay == 0 {
		e.ClearCombatInfo(c)
		return
	}
	e.stepSquare(c, int(ci.Facing))
	if c.Okay == 0 {
		e.ClearCombatInfo(c)
	}
	e.inClouds(c)
}

// AiAttackLoop = ovr10:0B3C (char): move, pick a target and attack until
// the turn is over; true when it is.
func (e *RuleState) AiAttackLoop(c *Character) bool {
	ci := c.Combat
	e.aiLastDir, e.aiBlocked, e.aiState3 = 8, 0, 0
	iter := 0
	cont, stop := true, false
	if e.Rules().Special(e, c) {
		e.AffectGroup(c, 0xE)
	}
	if c.Side == SideParty && e.Bandage(true) { // CotAB ovr010 sub_35DB1: a party member under the computer bandages first
		ci.Delay = 0
	}
	if ci.Delay == 0 {
		cont = false
	}
	for !stop && cont {
		for ci.Fleeing != 0 && ci.Moves > 0 && int8(ci.Delay) > 0 && int8(ci.Delay) < 0x14 {
			e.aiMoveStep(c)
		}
		if ci.Delay == 0 || ci.Delay == 0x14 {
			cont = false
		} else {
			stop = false
		}
		if stop || !cont {
			continue
		}
		iter++
		if iter > 0x14 {
			stop = true
			cont = !e.aiGiveUp(c)
		}
		engaged := false
		rng := 1
		if w := e.TurnWield(c).Weapon; w != nil {
			rng = int(e.Base(w.Type).Range - 1)
		}
		if rng == 0 || rng == 0xFF {
			rng = 1
		}
		tgt := ci.Target
		if tgt != nil && (tgt.Okay == 0 || tgt.Side == SideParty) {
			tgt = nil
		}
		if tgt != nil && e.MayActAgainst(c, tgt) {
			tx, ty := e.CombatX(tgt), e.CombatY(tgt)
			r := rng
			if e.LineOfSight(&r, &tx, &ty, e.CombatX(c), e.CombatY(c)) && r>>1 <= rng {
				engaged = true
			}
		}
		if !engaged {
			n := e.NearEnemies(c, rng)
			if n == 0 {
				if e.pickTarget(c, 0xFF, 0, 0) {
					e.aiMoveStep(c)
				} else {
					stop = e.aiGiveUp(c)
				}
			} else {
				foes := make([]*Character, 0, n)
				for i := 1; i <= n; i++ {
					foes = append(foes, e.Combatant(byte(e.Targets[i])))
				}
				if tgt = e.Rules().Target(e, c, foes); tgt == nil { // extra: the rule set's pick
					tgt = e.Combatant(byte(e.Targets[e.Dice(1, n)]))
				}
				if w := e.TurnWield(c); e.RangedWeapon(w) && !e.Throwable(w) && e.NearEnemies(c, 1) > 0 {
					e.AiPickWeapon(c)
					stop = true
					if w := e.TurnWield(c); e.RangedWeapon(w) && !e.Throwable(w) { // the swap was refused (cursed): the round ends (the Pascal's fix of a soft lock)
						stop = e.aiGiveUp(c)
					}
					continue
				}
				if e.RangeTo(c, tgt) == 1 || e.MayActAgainst(c, tgt) {
					engaged = true
				}
			}
		}
		if !engaged {
			continue
		}
		if e.Sweep(c, tgt) {
			stop = e.ClearCombatInfo(c)
			continue
		}
		e.CountAttackTurn(c, tgt)
		var weapon *Item
		with := e.TurnWield(c)
		if e.RangedWeapon(with) {
			e.AmmoFor(c, with, &weapon)
			if e.Throwable(with) && e.RangeTo(c, tgt) == 1 {
				weapon = nil
			}
		}
		stop = e.FullAttack(c, tgt, with, 0, weapon)
		if stop {
			cont = false
		} else if tgt.Okay == 0 {
			stop = true
		}
	}
	return !cont
}

// aiGiveUp = ovr10:0F4F (char): the turn ends; a combatant that cannot go
// on is cleared, otherwise 00BA:2686 finishes it.
func (e *RuleState) aiGiveUp(c *Character) bool {
	if e.IsHelpless(c) || e.RangedWeapon(e.TurnWield(c)) || c.Combat.Delay == 0 {
		return e.ClearCombatInfo(c)
	}
	return e.Guard(c)
}

// AiKey is the key the player pressed during a computer turn (ovr10:0FC1).
type AiKey byte

const (
	AiKeyNone       AiKey = iota
	AiKeyQuickMagic       // Alt-M: the computer casts for the party, or stops
	AiKeyTakeOver         // Space: every player character back under manual control
	aiKeyCheat            // Alt-X: the STING cheat, off without its parameter
)

// AiKeys = ovr10:0FC1 (char): the pending key at one of the computer turn's
// polls, taken (the original discarded pending keys). True when Space put c
// itself back under manual control: it acts again at once (initiative 20).
func (e *RuleState) AiKeys(c *Character) bool {
	k := e.Fight.Key
	e.Fight.Key = AiKeyNone
	switch k {
	case AiKeyQuickMagic:
		e.MagicOn ^= 1
	case AiKeyTakeOver:
		for _, p := range e.Party {
			if p.Control < 0x80 && p.Status != StatusAnimated {
				p.Auto = 0
			}
		}
		if c.Auto == 0 {
			c.Combat.Delay = 0x14
			return true
		}
	}
	return false
}

// PartyAuto: the computer fights for a member the player may take back (Quick), so its turns poll aiKeys.
func (e *RuleState) PartyAuto() bool {
	for _, p := range e.Party {
		if p.Auto != 0 && p.Control < 0x80 {
			return true
		}
	}
	return false
}

// aiFleeCheck = ovr10:10D7 (char): a forced flee, or a morale check that
// makes the combatant flee or surrender.
func (e *RuleState) aiFleeCheck(c *Character) bool {
	ci := c.Combat
	r := false
	ci.Fleeing = 0
	e.removeTiedAffects(c)
	if ci.Turned != 0 {
		ci.Fleeing = 1
		e.CombatSay(c, "is forced to flee", true)
		return r
	}
	if c.Control <= 0x7F || e.Rules().Holds(e, c) {
		return r
	}
	e.fleeRoll = byte(int(c.Control&0x7F) * 2)
	if e.fleeRoll > 0x66 && !e.T().MoraleUncapped { // PoD keeps it (pod-re combatai.pas CheckMorale)
		e.fleeRoll = 0
	}
	e.AffectGroup(c, 0x11)
	t := 100 - int(c.HP)*100/int(c.MaxHP) // cdq clobbers dx: signed word divide
	if !(int(e.fleeRoll) < t || e.fleeRoll == 0) {
		return r
	}
	e.fleeRoll = e.FoesHealth
	e.AffectGroup(c, 0x11)
	t2 := uint16(100 - e.Terms.Morale)
	if !(uint16(e.fleeRoll) < t2 || e.fleeRoll == 0 || c.Side == SideParty) {
		return r
	}
	a := e.fastestFoeMove(c)
	if a <= e.MoveRate(c)>>1 {
		ci.Fleeing = 1
		e.RemoveAffect(c, affStrikeTied, nil)
		e.RemoveAffect(c, affBloodDrain, nil)
		return r
	}
	if c.Stats[StatInt] > 5 {
		e.KillCharacter(c, StatusUnconscious, "Surrenders")
		r = e.ClearCombatInfo(c)
	}
	return r
}

// weaponScore = ovr10:128D (char, item): how good the item is as the
// character's weapon (0 = unusable).
func (e *RuleState) weaponScore(c *Character, it *Item) int {
	t := e.Base(it.Type)
	s := byte(int(t.Dice) * int(t.Sides))
	if int8(it.Plus) > 0 {
		s += byte(int(int8(it.Plus)) << 3)
	}
	if int8(t.Bonus) > 0 {
		s += byte(int(int8(t.Bonus)) << 1)
	}
	if e.isKind(it.Type, ItemIDHolyWater) && c.Combat.Target != nil && int8(c.Combat.Target.Undead) > 0 {
		s = 8
	}
	if t.Usage&8 != 0 {
		s += byte((int(t.Speed) - 1) << 1)
	}
	if t.Hands <= 1 {
		s += 3
	}
	if int(c.Hands)+int(t.Hands) > 3 {
		s = 0
	}
	if it.Spells[2] == ItemPower(affCursedAlignment) && byte(it.Spells[1]&0xF) != c.Alignment {
		s = 0
	}
	if it.Spells[1] == 0x53 {
		s = 0
	}
	if it.Cursed != 0 {
		s = 0
	}
	return int(s)
}

// AiPickWeapon = ovr10:13CA (char): choose the best melee or missile
// weapon, and the best second item, and ready them.
func (e *RuleState) AiPickWeapon(c *Character) {
	e.saveLoadout(c)
	e.ChooseWeapon(c, false) // extra: chosen anew, while the turn has not struck
	b := func(it *Item) *ItemBase { return e.Base(it.Type) }
	if w := c.Slots[SlotWeapon]; w != nil {
		c.Hands -= b(w).Hands
	}
	if w := c.Slots[SlotShield]; w != nil {
		c.Hands -= b(w).Hands
	}
	var best1, best2, best3 *Item
	bs1, bs3 := byte(1), byte(0)
	bs2 := byte(int(c.BaseDice[0]) * int(c.BaseSides[0]))
	if int8(c.BaseBonus[0]) > 0 {
		bs2 += byte(int(int8(c.BaseBonus[0])) << 1)
	}
	for _, p := range e.Carried(c) {
		if b(p).Slot == 0 && b(p).Classes&c.ClassMask != 0 {
			sc := byte(e.weaponScore(c, p))
			f := b(p).Usage
			if (f&8 != 0 || f&0x10 != 0) && sc > bs1 {
				best1, bs1 = p, sc
			}
			if f&8 == 0 && sc > bs2 {
				best2, bs2 = p, sc
			}
		}
		if b(p).Slot == 1 && b(p).Classes&c.ClassMask > 0 {
			sc := byte(0)
			if int8(p.Plus) >= 0 {
				sc = byte(int(int8(p.Plus)) + 1)
			}
			if sc > bs3 {
				best3, bs3 = p, sc
			}
		}
	}
	ranged := best1 != nil && b(best1).Range > 1 && b(best1).Usage&0x14 == 0x14
	var pend *Item
	var f byte
	if best1 != nil {
		f = b(best1).Usage
		if f&0x10 != 0 {
			pend = best1
		}
		if f&8 != 0 {
			if f&1 != 0 && c.Slots[SlotArrows] != nil && e.isKind(c.Slots[SlotArrows].Type, ItemIDArrow) { // not a missile in the slot
				pend = c.Slots[SlotArrows]
			}
			if f&0x80 != 0 {
				pend = c.Slots[SlotBolts]
			}
		}
	}
	ok := pend != nil || f == 0x0A
	chosen := best2
	if best1 != nil && int(bs1) > int(bs2)>>1 && ok && (ranged || e.NearEnemies(c, 1) == 0) {
		chosen = best1
	}
	if chosen != nil && e.aiDrawRanged(c, chosen) {
		return
	}
	done := false
	ok = true
	if w := c.Slots[SlotWeapon]; w != nil && (w == chosen || w.Cursed != 0) {
		ok = false
	}
	if chosen != nil && chosen.Readied != 0 { // extra: already readied (the off hand): readyItem would unready it
		ok = false
		chosen = nil
	}
	if ok {
		if c.Slots[SlotWeapon] != nil {
			e.ReadyItem(c.Slots[SlotWeapon])
		}
		e.ApplyItems(c)
		if w := c.Slots[SlotShield]; w != nil && w.Cursed == 0 {
			c.Hands -= b(w).Hands
		}
		if chosen != nil {
			e.ReadyItem(chosen)
		}
		done = true
	}
	e.ApplyItems(c)
	e.SetSpeed(c, e.Hand(c))
	ok = true
	if w := c.Slots[SlotShield]; w != nil && (w == best3 || w.Cursed != 0) {
		ok = false
	}
	if c.Hands > 2 {
		// the original passes chosen even when nil
		if w := c.Slots[SlotShield]; w == nil || w.Cursed != 0 {
			if chosen != nil {
				e.ReadyItem(chosen)
			}
		} else {
			e.ReadyItem(w)
		}
		done = true
	} else if c.Hands < 2 && ok {
		if c.Slots[SlotShield] != nil {
			e.ReadyItem(c.Slots[SlotShield])
		}
		e.ApplyItems(c)
		if best3 != nil {
			e.ReadyItem(best3)
		}
		done = true
	}
	e.ApplyItems(c)
	if done {
	}
}

// aiDrawRanged (extra): a chosen launcher or thrown-only weapon readies
// into its own slot (no hands: the weapon in hand and the shield stay) and
// is what the turn's attacks are made with (ChooseWeapon); false for a
// weapon of the hand.
func (e *RuleState) aiDrawRanged(c *Character, chosen *Item) bool {
	if s := e.readySlot(chosen.Type); s != SlotLauncher && s != SlotMissiles {
		return false
	}
	if chosen.Readied == 0 {
		if o := e.InTheWay(c, chosen); o != nil {
			e.ReadyItem(o)
			e.ApplyItems(c)
		}
		e.ReadyItem(chosen)
	}
	e.ApplyItems(c)
	e.ChooseWeapon(c, true)
	return true
}

// AiPathDir (extra) is the first step of the cheapest walk (A*, in
// moves: 2 a step, 3 diagonally, times the ground's cost) that brings c
// next to t around walls and other combatants, or -1 when there is none
// or c is next to t already. The move budget is not counted: the walk may
// take more than one round.
func (e *RuleState) AiPathDir(c, t *Character) int {
	if t == nil {
		return -1
	}
	tx, ty, n := e.CombatX(t), e.CombatY(t), e.CombatSize(t)
	next := func(x, y int) bool { return x >= tx-1 && x <= tx+n && y >= ty-1 && y <= ty+n }
	h := func(x, y int) int { // octile distance to the squares next to t
		dx := max(tx-1-x, x-tx-n, 0)
		dy := max(ty-1-y, y-ty-n, 0)
		return 2*max(dx, dy) + min(dx, dy)
	}
	idx := e.CombatIndex(c)
	ox, oy := e.Cpos[idx].X, e.Cpos[idx].Y
	defer func() { e.Cpos[idx].X, e.Cpos[idx].Y = ox, oy }()
	sx, sy := int(ox), int(oy)
	if next(sx, sy) {
		return -1
	}
	var cost [25][50]int   // moves to get there, +1 (0: not reached)
	var first [25][50]int8 // the first step on the way there
	cost[sy][sx] = 1
	open := &e.aiOpen
	open.reset(h(sx, sy))
	open.push(aiNode{int32(sx), int32(sy), int32(h(sx, sy)), int32(h(sx, sy))})
	for open.n > 0 {
		n := open.pop()
		at := struct{ x, y, f, h int }{int(n.x), int(n.y), int(n.f), int(n.h)}
		if next(at.x, at.y) {
			return int(first[at.y][at.x])
		}
		g := cost[at.y][at.x]
		if at.f-at.h != g-1 {
			continue // a cheaper way here was found after this one was queued
		}
		e.Cpos[idx].X, e.Cpos[idx].Y = byte(at.x), byte(at.y)
		for d := 0; d < 8; d++ {
			x, y := at.x+DirDX[d], at.y+DirDY[d]
			if x < 0 || x >= 50 || y < 0 || y >= 25 {
				continue
			}
			who, gr := e.GroundAt(c, d)
			gc := int(e.GroundInfo(Ground(gr))[0])
			if who > 0 || gr == 0 || gc == 0xFF {
				continue
			}
			ng := g + (2+d&1)*max(gc, 1)
			if cost[y][x] != 0 && cost[y][x] <= ng {
				continue
			}
			cost[y][x] = ng
			if at.x == sx && at.y == sy {
				first[y][x] = int8(d)
			} else {
				first[y][x] = first[at.y][at.x]
			}
			hx := h(x, y)
			open.push(aiNode{int32(x), int32(y), int32(ng - 1 + hx), int32(hx)})
		}
	}
	return -1
}

// aiNode is a square on aiPathDir's open list, f its moves so far plus
// the least it can still take, h that least.
type aiNode struct{ x, y, f, h int32 }

// aiOpen is aiPathDir's open list: a ring of buckets by f (Dial). A step
// raises f by at most twice its cost, 6 on the ground costs there are (1,
// or 0xFF for walls), so the ring holds every f still open.
type aiOpen struct {
	b [32][]aiNode
	f int // the lowest bucket that may hold a node
	n int // nodes held
}

func (o *aiOpen) reset(f int) {
	for i := range o.b {
		o.b[i] = o.b[i][:0]
	}
	o.f, o.n = f, 0
}

func (o *aiOpen) push(n aiNode) {
	b := &o.b[int(n.f)%len(o.b)]
	*b = append(*b, n)
	o.n++
}

// pop takes a node of the lowest f, the latest pushed of them first (the
// one furthest along, as a rule).
func (o *aiOpen) pop() aiNode {
	for len(o.b[o.f%len(o.b)]) == 0 {
		o.f++
	}
	b := &o.b[o.f%len(o.b)]
	n := (*b)[len(*b)-1]
	*b = (*b)[:len(*b)-1]
	o.n--
	return n
}
