package rules1e

import (
	"cmp"
	"gobox/rules"
	"slices"
)

// BeginTurn = por takeTurn's opening (ovr09:01F6): the turn's counters
// cleared, the rule set's TurnStart (extra), affect group 7; Skipped when
// the actor's delay is spent; else c becomes the actor (Who, as the engine
// selects it), its items applied (ApplyItems) and affect group 0xF run.
// The actor before the skip check is whoever the engine selected.
func (r *Rules) BeginTurn(_ rules.Fight, id rules.CharID, _ *rules.RNG) rules.Turn {
	c := r.chars[id]
	if c == nil {
		return rules.Turn{Skipped: true}
	}
	ci := c.Combat
	ci.Attackers, ci.Flank, ci.Guarding = 0, 0, 0
	r.S.Rules().TurnStart(r.S, c)
	r.S.AffectGroup(c, 7)
	if int8(ci.Delay) <= 0 {
		return rules.Turn{Skipped: true}
	}
	if ci.Delay == 0x14 {
		ci.Delay = 0x13
	}
	if i := r.S.PartyIndex(c); i < len(r.S.Party) { // the engine selects c now
		r.S.Who = i
	}
	r.S.ApplyItems(c)
	r.S.AffectGroup(c, 0xF)
	return rules.Turn{Computer: c.Auto != 0}
}

// autoTurn = por computerTurn for the Fight's Actor.
func (r *Rules) autoTurn(sit rules.Situation, _ rules.AutoTurn) (rules.Outcome, error) {
	f, ok := sit.(rules.Fight)
	if !ok {
		return rules.Outcome{}, errNotFight
	}
	c := r.chars[f.Actor]
	if c == nil {
		return rules.Outcome{}, errNoChar
	}
	return r.S.ComputerTurn(c), nil
}

// BeginFight = por combatSetup's rules opening (ovr11:1E31): the rule set's
// CombatStart, the last fight's state cleared. The words it reads are f.Terms
// (the engine capped the morale).
func (r *Rules) BeginFight(f rules.Fight, _ *rules.RNG) rules.Outcome {
	s := r.S
	s.Fight = FightState{On: true} // por pushCombat's, before combatSetup
	r.straight, s.Terms = f.Board.Straight, f.Terms
	s.PosX, s.PosY, s.Facing = f.Terms.PartyAt.X, f.Terms.PartyAt.Y, int(f.Terms.PartyFacing) // the Engine's, given for the fight
	s.Rules().CombatStart(s)
	s.MagicOn, s.CombatRound, s.Rv.HitRoll = 0, 0, 0
	s.Nbodies = 0
	s.Clouds = nil
	s.ThrownCopy = nil
	s.Bodies = [10]Body{}
	s.AttackedAlly = 0
	return rules.Outcome{}
}

// fightPlaced = the end of por combatSetup: affect group 8 on every
// combatant, the foes' health taken. The actor (Who) is the engine's.
func (r *Rules) fightPlaced() {
	s := r.S
	for _, c := range s.Party {
		s.AffectGroup(c, grpRound)
		s.AffectGroup(c, grpAlwaysInvisible)
	}
	s.FoesHealthPercent()
	for _, c := range s.Party {
		s.AffectGroup(c, grpCombatStart)
	}
}

// EndFight = por AfterCombat's rules part (ovr06): combatResult (who won,
// the experience given, the fallen and the fled dealt with), then
// cleanupCombatants (the monsters leave, the beaten are counted,
// the members drop their combat blocks). The actor (Who) is the engine's
// selected member (a duel's experience is its level).
func (r *Rules) EndFight(f rules.Fight, _ *rules.RNG) rules.FightResult {
	s := r.S
	s.Terms = f.Terms
	defer func() { s.Terms = rules.FightTerms{} }() // the words are the fight's: out of it a bonus is none
	if s.ageCheckDue {
		s.oldAgeDeaths()
	}
	res := r.combatResult(f.Terms.Duel)
	for i := 0; i < len(s.Party); {
		c := s.Party[i]
		extra := c.Combat != nil && c.Combat.Extra == 1
		if extra || c.Side == SideMonsters {
			if c.Okay != 1 && c.Status != StatusReserve {
				res.Beaten++
			}
			r.remove(i, extra)
			continue
		}
		c.Combat = nil
		// extra: the fight's Quick ends with it (as Manual, AiKeyTakeOver)
		if c.Control < 0x80 && c.Status != StatusAnimated {
			c.Auto = 0
		}
		i++
	}
	res.AttackedAlly = s.AttackedAlly != 0
	if res.Verdict != rules.PartyLost || f.Terms.Duel {
		res.Shared, res.Sharers = r.HiredShare()
	}
	res.Money = s.Loot
	for _, it := range s.Treasure {
		res.Items = append(res.Items, r.Handle(it))
	}
	return res
}

// HiredShare = ovr06:12E8's split: hired NPCs (+0x84 > 0x7F) take their
// shares (+0x85 & 7) of the money before the party gets to it. Money is
// copper (Felix's ruling): the coins are cut as one copper pile, where the
// original cut each coin type and left each type's remainder
// (CHANGES-FROM-ORIGINAL.md). Gems and jewelry are counted pieces.
// shared: a share was taken (the page shows); sharers: who took one.
func (r *Rules) HiredShare() (shared bool, sharers []rules.CharID) {
	s := r.S
	shares, parts := 0, 0
	for _, c := range s.Party {
		if c.Control > 0x7F && c.Status == StatusOkay {
			shares += int(c.Share & 7)
			parts += int(c.Share & 7)
		} else {
			parts++
		}
	}
	if shares == 0 || s.Loot.Empty() {
		return false, nil
	}
	rates := rules.GoldBoxRates
	copper := s.Loot.Gold*rates.CopperPerGold + s.Loot.Copper
	copper -= copper / uint64(parts) * uint64(shares)
	s.Loot.Gold, s.Loot.Copper = copper/rates.CopperPerGold, copper%rates.CopperPerGold
	for _, n := range []*uint64{&s.Loot.Gems, &s.Loot.Jewelry} {
		*n -= *n / uint64(parts) * uint64(shares)
	}
	for _, c := range s.Party {
		if c.Control > 0x7F && c.Status == StatusOkay && c.Share > 0 {
			sharers = append(sharers, r.ID(c))
		}
	}
	return true, sharers
}

// remove takes member i out of the party (por RemoveSelected's rules part);
// keepCount: an extra, never counted in Game.PartyCount.
func (r *Rules) remove(i int, keepCount bool) {
	r.Dismiss(r.ID(r.S.Party[i]))
	if !keepCount {
		r.S.Game.PartyCount--
	}
}

// combatResult = ovr06:04C9: who won, experience, and what happens to the
// fallen. Verdict PartyLost is the original's quit (DS:49FF), the whole
// party lost; a duel is never quit, a lost duel is PartyLost too.
func (r *Rules) combatResult(duel bool) rules.FightResult {
	s := r.S
	var res rules.FightResult
	idle := 0
	quit, fled, won := true, false, false
	for _, c := range s.CombatParty() {
		if c.Status == StatusRunning {
			fled = true
		}
	}
	if duel {
		quit = false
	}
	if !duel {
		for _, c := range s.CombatParty() {
			st := c.Status
			if (st == StatusRunning || st == StatusAnimated || st == StatusOkay || s.Terms.DuelOpponent == 1) &&
				c.Side == SideParty && c.Control < 0x80 {
				quit = false
			}
			if st == StatusAnimated || st == StatusOkay {
				won, fled = true, false
			}
			if c.Okay == 0 || st == StatusAnimated {
				idle++
			}
		}
		if won {
			res.Exp, res.Foes = r.computeExp(duel, idle)
			s.GiveExp(res.Exp)
		}
		if !quit {
			if !fled {
				for _, c := range s.CombatParty() {
					switch st := c.Status; {
					case st == StatusRunning:
						c.Status, c.Okay = StatusOkay, 1
					case st == StatusDying:
						c.Status = StatusUnconscious
					case st == StatusUnconscious && c.HP > 0:
						c.Status, c.Okay = StatusOkay, 1
					}
				}
			} else {
				for i := 0; i < len(s.CombatParty()); {
					c := s.Party[i]
					if c.Status == StatusRunning {
						c.Status, c.Okay = StatusOkay, 1
						i++
						continue
					}
					r.remove(i, false)
				}
			}
		} else {
			for n := len(s.CombatParty()); n > 0; n-- {
				r.remove(0, false)
			}
			s.Game.PartyCount = 0
		}
		res.Verdict = verdict(quit, fled, won)
		return res
	}
	for _, c := range s.Party {
		if c.Okay == 1 && c.Status == StatusOkay && c.Side != SideMonsters {
			won = true
			var foes bool
			res.Exp, foes = r.computeExp(duel, idle)
			res.Foes = res.Foes || foes
			s.GiveExp(res.Exp)
		}
	}
	for _, c := range s.Party {
		switch c.Status {
		case StatusOkay, StatusAnimated:
			c.Okay = 1
		case StatusDying:
			c.Status = StatusUnconscious
		}
	}
	res.Verdict = verdict(!won, false, won) // a duel: won or lost
	return res
}

// verdict names combatResult's flags; lost (quit) goes first, a win ends a flight.
func verdict(lost, fled, won bool) rules.Verdict {
	switch {
	case lost:
		return rules.PartyLost
	case fled:
		return rules.PartyFled
	case won:
		return rules.PartyWon
	}
	return rules.FightOn
}

// computeExp = ovr06:0000 (var exp): the beaten monsters' worth, their
// money and (some of) their items go to the treasure, and the treasure's
// worth counts too; the sum is split over the members that earn it.
// foes: there were monsters to beat.
func (r *Rules) computeExp(duel bool, idle int) (exp int32, foes bool) {
	s := r.S
	if duel && s.Terms.DuelOpponent != 0 {
		sel := &Character{}
		if len(s.Party) > 0 {
			sel = s.Party[s.Who]
		}
		return int32(int16(int8(sel.MaxLevel))) * 100, false
	}
	var total int32
	taken := 0
	for _, c := range s.Party {
		if c.Side != SideMonsters || c.Status == StatusRunning || c.Status == StatusReserve { // pod-re postcbt.pas:97: the reserve was not beaten
			continue
		}
		foes = true
		var coins [7]uint16
		for i := range coins {
			coins[i] = c.Coins[i]
		}
		s.Loot.Add(CoinsToPurse(coins), rules.GoldBoxRates)
		// the byte product is a 16-bit mul, sign-extended
		total += int32(c.ExpValue) + int32(int16(uint16(c.HPRoll)*uint16(c.ExpPerHP)))
		if s.Terms.NoItemLoot {
			continue
		}
		for _, it := range c.Items {
			if it.Price != 0 || taken < 8 && s.Dice(1, 10) <= 3 {
				taken++
				s.ItemName(c, it, false)
				cp := *it
				cp.Readied = 0
				s.Treasure = append([]*Item{&cp}, s.Treasure...)
			}
		}
	}
	total += int32(s.Loot.Gold + s.Loot.Gems*250 + s.Loot.Jewelry*2200)
	// (the original stops at DS:5D9A, an item the combat code marks; it is
	// never set outside a fight)
	for _, it := range s.Treasure {
		if p := int8(it.Plus); p > 0 {
			total += int32(p) * 400
		}
	}
	// ponytail: the original divides by zero when d is 0
	if d := int32(uint16(s.Game.PartyCount - uint16(idle))); d > 0 {
		total /= d
	}
	return total, foes
}

// turnVerb = por turnChoice's verbs that take no square, for the Fight's
// Actor (Done's Guard, Wait and Quit, Turn, Bandage, Quick). Quick puts c
// (All: every player character too) on Quick; the engine then plays the
// computer turn (Decide).
func (r *Rules) turnVerb(sit rules.Situation, a rules.Action) (rules.Outcome, error) {
	f, ok := sit.(rules.Fight)
	if !ok {
		return rules.Outcome{}, errNotFight
	}
	c := r.chars[f.Actor]
	if c == nil {
		return rules.Outcome{}, errNoChar
	}
	s := r.S
	var done bool
	switch a := a.(type) {
	case rules.Guard:
		done = s.Guard(c)
	case rules.Move: // a walk to a square MoveReach offers (the actor's own is none)
		at := [2]int{int(s.Cpos[s.CombatIndex(c)].X), int(s.Cpos[s.CombatIndex(c)].Y)}
		to := [2]int{a.To.X, a.To.Y}
		if c.Combat.Moves > 0 && to != at {
			if p, ok := s.MoveReach(c)[to]; ok {
				done = s.Walk(c, p)
			}
		}
	case rules.Flee: // moveAttack (ovr09:09DF): a step off the map, asked first
		if s.FleeDir(c) >= 0 {
			return rules.Outcome{Confirm: "Flee:"}, nil // ovr09:0C17
		}
	case rules.Attack:
		return r.attack(c, a, false)
	case rules.Confirmed:
		switch q := a.Action.(type) {
		case rules.Flee: // answered Yes
			if d := s.FleeDir(c); d >= 0 {
				s.TurnTo(c, byte(d))
				done = s.TryFlee(c)
			}
		case rules.Attack:
			return r.attack(c, q, true)
		}
	case rules.Delay: // Done's Wait
		c.Combat.Delay = 1
		done = true
	case rules.EndTurn: // Done's Quit
		done = s.ClearCombatInfo(c)
	case rules.TurnUndead:
		s.ClericTurnUndead(c)
		done = s.ClearCombatInfo(c)
	case rules.Bandage:
		s.Bandage(true)
		done = s.ClearCombatInfo(c)
	case rules.Quick:
		s.QuickOne(c)
		if a.All { // extra: Shift-U, the whole party
			for _, p := range s.Party {
				if p.Control < 0x80 {
					s.QuickOne(p)
				}
			}
		}
		done = true
	}
	return rules.Outcome{TurnOver: done}, nil
}

// attack = por aimConfirm/aimTake's blow (ovr14:2C3D, mode 1): an ally
// asks "Attack Ally?" first (sure: answered Yes); a sweep, or the full
// attack with the ammunition it takes. TurnOver: the attack was made.
func (r *Rules) attack(c *Character, a rules.Attack, sure bool) (rules.Outcome, error) {
	s := r.S
	t := r.chars[a.Target]
	if t == nil {
		return rules.Outcome{}, errNoChar
	}
	if !s.ChooseWeapon(c, a.Ranged) { // extra: the turn's weapon, the hand's or (the Ranged action) the Launcher or Missiles slot's; it stands once the turn has struck
		return rules.Outcome{}, ErrNotNow
	}
	w := s.TurnWield(c)
	if s.AsksAlly(c, t) {
		if !sure {
			return rules.Outcome{Confirm: "Attack Ally?"}, nil
		}
		s.AllyAttacked()
	}
	if s.Sweep(c, t) {
		return rules.Outcome{TurnOver: s.ClearCombatInfo(c)}, nil
	}
	var res *Item
	s.CountAttackTurn(c, t)
	if s.RangedWeapon(w) && s.AmmoFor(c, w, &res) && s.Throwable(w) && s.RangeTo(c, t) == 1 {
		res = nil
	}
	return rules.Outcome{TurnOver: s.FullAttack(c, t, w, 0, res)}, nil
}

// aims hands the player's aims to the cast (Fight.Aims, aimNow takes them).
func (r *Rules) aims(as []rules.Aim) {
	r.S.Fight.Aims = nil
	for _, a := range as {
		r.S.Fight.Aims = append(r.S.Fight.Aims, AimAnswer{Found: a.Found,
			T: AimTarget{C: r.chars[a.Target], X: byte(a.At.X), Y: byte(a.At.Y)}})
	}
}

// cast = por castRun's cwInfo (Begun: the spell begun earlier goes off, the
// turn is gone all the same) and cwTurn (the turn menu's Cast) for the
// Fight's Actor, with the aims taken first. TurnOver: castTurnSpell's ok.
func (r *Rules) cast(sit rules.Situation, a rules.Cast) (rules.Outcome, error) {
	f, ok := sit.(rules.Fight)
	if !ok {
		return rules.Outcome{}, errNotFight
	}
	c := r.chars[f.Actor]
	n := slices.Index(SpellIDs[:], a.Spell)
	if c == nil {
		return rules.Outcome{}, errNoChar
	}
	if n <= 0 {
		return rules.Outcome{}, errBadValue
	}
	s := r.S
	r.aims(a.Aims)
	s.Fight.Choice = a.Choice
	defer func() { s.Fight.Aims, s.Fight.Choice = nil, 0 }()
	if a.Begun {
		var quit bool
		s.CastSpell(SpellID(n), 0, 1, CastReply{Fight: true}, &quit)
		s.ClearCombatInfo(c)
		return rules.Outcome{TurnOver: true}, nil
	}
	s.CastTurnSpell(n, 0, &ok)
	return rules.Outcome{TurnOver: ok}, nil
}

// useItem = por castRun's cwItem: the Held item casts its own spell
// (UsingItem) or the scroll's Spell with the aims taken first; the user's
// speed is set again. TurnOver: UseItem's quit.
func (r *Rules) useItem(sit rules.Situation, a rules.UseItem) (rules.Outcome, error) {
	c, it := r.chars[a.User], r.Held(a.Item)
	if c == nil || it == nil {
		return rules.Outcome{}, errNoChar
	}
	s := r.S
	own := !s.IsScroll(it)
	n := int(it.Spells[1])
	if !own {
		if n = slices.Index(SpellIDs[:], a.Spell); n <= 0 {
			return rules.Outcome{}, errBadValue
		}
		s.takeScroll(c, it)
	}
	r.aims(a.Aims)
	defer func() { s.Fight.Aims = nil }()
	s.UsingItem = own
	quit := s.UseItem(it, SpellID(n), CastReply{Fight: true})
	s.SetSpeed(c, s.Hand(c))
	return rules.Outcome{TurnOver: quit}, nil
}

// maneuver = por's maneuver verbs (extra, combatrules.go) for the Fight's
// Actor: a weapon maneuver is armed (Ats[0]: the square aimed to move to
// first), Cancel disarms it; the others are used on Target (an area one at
// Ats[0]), a self one ends the turn. A Target of the wrong side is ErrNotNow.
func (r *Rules) maneuver(sit rules.Situation, a rules.Maneuver) (rules.Outcome, error) {
	f, ok := sit.(rules.Fight)
	if !ok {
		return rules.Outcome{}, errNotFight
	}
	c := r.chars[f.Actor]
	if c == nil {
		return rules.Outcome{}, errNoChar
	}
	s := r.S
	if a.Cancel {
		if len(a.Ats) > 0 && s.At(c) != a.Ats[0] {
			s.MoveTo(c, a.Ats[0], false)
		}
		s.Fight.Armed = Armed{}
		return rules.Outcome{}, nil
	}
	id := int(a.ID)
	ms := s.Rules().Maneuvers(s, c)
	i := slices.IndexFunc(ms, func(m Maneuver) bool { return m.ID == id })
	if i < 0 {
		return rules.Outcome{}, ErrNotNow
	}
	m := ms[i]
	t := r.chars[a.Target]
	switch m.Aim {
	case AimWeapon:
		if len(a.Ats) > 0 {
			s.Fight.Aimed = a.Ats[0]
		}
		s.Rules().Use(s, c, id, nil) // what comes before its attack (a shift, the move aimed at)
		s.Fight.Aimed = rules.BoardPos{X: -1, Y: -1}
		s.Fight.Armed = Armed{C: c, ID: id}
		return rules.Outcome{}, nil
	case AimSelf:
		return rules.Outcome{TurnOver: s.Rules().Use(s, c, id, nil) && s.ClearCombatInfo(c)}, nil
	case AimArea:
		if len(a.Ats) > 0 {
			s.Fight.Aimed = a.Ats[0]
		}
	default:
		if t == nil || (t.Side == c.Side) != (m.Aim == AimAlly) {
			return rules.Outcome{}, ErrNotNow
		}
	}
	return rules.Outcome{TurnOver: s.Rules().Use(s, c, id, t)}, nil
}

// fightAllowed = por combatMenu's offer (ovr09:0666) for the Fight's Actor,
// in the menu's order: Attack, Guard (Done's), UseItem (any item carried),
// Cast{} (any of the 20 memorized) with a Cast{Spell} per spell castable in
// combat, a Maneuver per maneuver, TurnUndead, Bandage, Delay (Wait),
// EndTurn, Quick, Flee. A pure query: what the menu runs first is
// openTurn's (Resolve(core.OpenTurn)).
func (r *Rules) fightAllowed(f rules.Fight) rules.Allowed {
	c := r.chars[f.Actor]
	var a rules.Allowed
	if c == nil {
		return a
	}
	s := r.S
	offer := func(x rules.Action, name string) { a.Offers = append(a.Offers, rules.Offer{Action: x, Name: name}) }
	ci := c.Combat
	offer(rules.Attack{}, "")
	if w := s.Hand(c); !s.RangedWeapon(w) || s.Throwable(w) {
		offer(rules.Guard{}, "")
	}
	if len(s.Carried(c)) > 0 { // ovr09:06E1: Use with any item carried (char +0xC7; extra: the party's pool among them)
		offer(rules.UseItem{}, "")
	}
	if ci.CanCast != 0 && !s.Terms.NoMagic {
		if slices.ContainsFunc(c.Memorized[:0x14], func(v SpellID) bool { return v > 0 }) { // ovr09:0711: any of the 20 memorized (+0x17)
			offer(rules.Cast{}, "")
		}
		seen := map[SpellID]bool{}
		for _, v := range c.Memorized[:0x15] {
			if v > 0 && s.T().hasSpell(v) && !seen[v] && s.SpellData(int(v), 5) != 0 { // ruling U4: only spells castable in combat (whenCast <> 0); a number the game lacks has no spell
				seen[v] = true
				offer(rules.Cast{Caster: f.Actor, Spell: SpellIDs[v]}, s.SpellName(int(v)))
			}
		}
	}
	for _, m := range s.Rules().Maneuvers(s, c) {
		offer(rules.Maneuver{ID: rules.ManeuverID(m.ID)}, m.Name)
	}
	if int8(c.Levels[ClsCleric]) > 0 && ci.TurnUsed == 0 && s.canTurn(c) { // ovr09:0782: a cleric's level (+0x96), Turn not used (+0x11); extra (Felix's ruling): only with undead it can turn
		offer(rules.TurnUndead{}, "")
	}
	if s.Bandage(false) {
		offer(rules.Bandage{Who: f.Actor}, "")
	}
	if slices.ContainsFunc(s.Party, func(o *Character) bool { return o != c && int8(o.Combat.Delay) > 0 }) { // extra (Felix's ruling): Wait only with someone still to act this round (NextCombatant); the last one came straight back
		offer(rules.Delay{}, "")
	}
	offer(rules.EndTurn{}, "")
	offer(rules.Quick{}, "")
	if ci.Moves > 1 && s.FleeDir(c) >= 0 {
		offer(rules.Flee{}, "")
	}
	reach := r.fightReach(&a, c)
	r.fightTargets(&a, c, reach)
	return a
}

// fightReach = por turnChoices' walk: the squares MoveReach offers but the
// actor's own, each with its steps (WalkSteps), sorted by row then column.
func (r *Rules) fightReach(a *rules.Allowed, c *Character) map[[2]int][]int {
	s := r.S
	reach := map[[2]int][]int{}
	if c.Combat.Moves > 0 {
		reach = s.MoveReach(c)
		delete(reach, [2]int{s.CombatX(c), s.CombatY(c)})
	}
	for sq, p := range reach {
		a.Moves = append(a.Moves, r.walk(c, rules.BoardPos{X: sq[0], Y: sq[1]}, p))
	}
	slices.SortFunc(a.Moves, func(x, y rules.Reach) int {
		return cmp.Or(cmp.Compare(x.At.Y, y.At.Y), cmp.Compare(x.At.X, y.At.X))
	})
	return reach
}

func (r *Rules) walk(c *Character, to rules.BoardPos, p []int) rules.Reach {
	w := rules.Reach{At: to}
	for _, st := range r.S.WalkSteps(c, p) {
		w.Path = append(w.Path, st.At)
		w.Free = append(w.Free, st.Free)
	}
	return w
}

// fightTargets = por turnChoices' foes: an Attack on each the actor may act
// against that the readied weapon reaches or a walk brings it next to.
func (r *Rules) fightTargets(a *rules.Allowed, c *Character, reach map[[2]int][]int) {
	s := r.S
	for i := 1; i <= s.Ncomb; i++ {
		f := s.Combatant(byte(i))
		if f == nil || f.Side == c.Side || f.Okay == 0 || s.Cpos[i].Size == 0 || !s.MayActAgainst(c, f) {
			continue
		}
		p := s.PathNextTo(c, f, reach)
		if w := s.Hand(c); !s.CanTarget(c, f, w, s.WeaponRange(w), 1) && p == nil {
			continue
		}
		t := rules.Aimed{Action: rules.Attack{Target: r.ID(f)}, Targets: 1}
		n := s.CombatSize(f)
		for d := range n * n {
			t.Squares = append(t.Squares, rules.BoardPos{X: s.CombatX(f) + d%n, Y: s.CombatY(f) + d/n})
		}
		if p != nil {
			end := [2]int{s.CombatX(c), s.CombatY(c)}
			for _, d := range p {
				end = [2]int{end[0] + DirDX[d], end[1] + DirDY[d]}
			}
			w := r.walk(c, rules.BoardPos{X: end[0], Y: end[1]}, p)
			t.Walk = &w
		}
		a.Aimed = append(a.Aimed, t)
	}
}

// takeLoot = por TakeLoot's rule (ovr06:0C06): BuyItem false is taken, the
// piece leaves the treasure; All goes on until one does not fit.
func (r *Rules) takeLoot(a rules.TakeLoot) error {
	s := r.S
	c := r.chars[a.Who]
	i := slices.Index(s.Party, c)
	if c == nil || i < 0 {
		return errNoChar
	}
	s.Who = i
	if a.All {
		for len(s.Treasure) > 0 && !s.BuyItem(s.Treasure[0]) {
			s.Treasure = s.Treasure[1:]
		}
		return nil
	}
	it := r.items[a.Item]
	if it == nil || !slices.Contains(s.Treasure, it) {
		return ErrNotNow
	}
	if !s.BuyItem(it) {
		s.Treasure = slices.DeleteFunc(s.Treasure, func(x *Item) bool { return x == it })
	}
	return nil
}
