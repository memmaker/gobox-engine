package rules1e

import (
	"gobox/rules"
	"slices"
)

// The fight: the board, aims, units, loot and the turn (core/fight.go).

// checkFight answers the Checks of the fight topic.
func (r *Rules) checkFight(k rules.Check) (rules.Answer, bool) {
	s := r.S
	switch k := k.(type) {
	case rules.InFight:
		return rules.Answer{Yes: r.S.Fight.On}, true
	case rules.ArcDir:
		dir := 0
		for !s.InArc(k.Start.X, k.Start.Y, k.To.X, k.To.Y, dir) {
			dir++
		}
		return rules.Answer{Value: int32(dir)}, true
	case rules.UnitBody:
		c, p := s.Combatant(byte(k.Unit)), s.Cpos[k.Unit]
		return rules.Answer{
			Yes:   c.Combat != nil && c.Combat.Extra == 0,
			Value: int32(s.GroundSquare(byte(s.grounds().Body)).Tile),
			Low:   int32(p.X), High: int32(p.Y),
		}, true
	case rules.BoardUp:
		return rules.Answer{Yes: r.S.Cmap != nil}, true
	case rules.FightRound:
		return rules.Answer{Yes: true, Value: int32(r.S.CombatRound)}, true
	default:
		return rules.Answer{}, false
	}
}

// resolveFight answers the Actions of the fight topic.
func (r *Rules) resolveFight(sit rules.Situation, a rules.Action) (rules.Outcome, bool, error) {
	s := r.S
	switch a := a.(type) {
	case rules.OpenTurn:
		return rules.Outcome{}, true, r.openTurn(a)
	case rules.CombatSay:
		s.CombatSay(r.chars[a.Who], a.Text, true)
		return rules.Outcome{}, true, nil
	case rules.AutoTurn:
		o, err := r.autoTurn(sit, a)
		return o, true, err
	case rules.Guard, rules.Delay, rules.EndTurn, rules.TurnUndead, rules.Quick, rules.Move, rules.Flee, rules.Attack:
		o, err := r.turnVerb(sit, a)
		return o, true, err
	case rules.Confirmed:
		switch a.Action.(type) {
		case rules.Flee, rules.Attack:
			o, err := r.turnVerb(sit, a)
			return o, true, err
		}
		return rules.Outcome{}, false, nil
	case rules.Bandage:
		if _, ok := sit.(rules.Fight); ok {
			o, err := r.turnVerb(sit, a)
			return o, true, err
		}
		return rules.Outcome{}, false, nil
	case rules.DetectLoot:
		return rules.Outcome{}, true, r.detectLoot()
	case rules.FightCleared:
		r.fightCleared()
	case rules.FightPlaced:
		r.fightPlaced()
	case rules.TakeLoot:
		return rules.Outcome{}, true, r.takeLoot(a)
	case rules.BankLoot:
		s.MoneySound(s.Loot)
		s.Purse.Add(s.Loot, rules.GoldBoxRates)
		s.Loot = Purse{}
	case rules.BuildGround:
		r.buildGround(a)
	case rules.PlaceFight:
		r.combatBlocks()
		return rules.Outcome{Effects: r.placeCombatants(a.Walls)}, true, nil
	case rules.ClearLoot:
		s.ClearLoot() // DS:67F4
		s.Treasure = nil
	case rules.ClearTreasure:
		s.Treasure = nil
	case rules.DevTreasure:
		s.Treasure = s.devItems(ItemGauntlets)
	case rules.AddTreasureBlock:
		b := r.TreasureBlock(int(a.Block))
		if len(b) == 0 {
			return rules.Outcome{}, true, errNoBlock
		}
		b, err := s.gameItems(b) // the block is the game's
		if err != nil {
			return rules.Outcome{}, true, err
		}
		s.AddTreasureBlock(b)
	case rules.AddTreasureRecords:
		return rules.Outcome{}, true, r.addTreasureRecords(a)
	case rules.SetSpeed:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		s.SetSpeed(c, s.Hand(c))
	case rules.SetLoot:
		s.SetLootCoins(a.Coins)
	case rules.AddLoot:
		s.Loot.Add(a.Loot, rules.GoldBoxRates)
	case rules.AddLibraryItems:
		return rules.Outcome{}, true, s.AddLibraryItems(a.IDs, a.Known)
	case rules.AddRandomTreasure:
		if err := s.AddRandomTreasure(int(a.N)); err != nil {
			return rules.Outcome{}, true, err
		}
		c := r.chars[a.Who]
		for _, it := range s.Treasure {
			s.ItemName(c, it, false) // 00BA:0442
		}
	case rules.Rob:
		robbed := r.party(a.Robbed)
		share := float64(a.Pct) / 100 * float64(len(robbed)) / float64(max(len(s.Party), 1))
		for _, n := range []*uint64{&s.Purse.Gold, &s.Purse.Copper, &s.Purse.Gems, &s.Purse.Jewelry} {
			*n -= uint64(float64(*n) * share)
		}
		for _, c := range robbed {
			s.ItemSave(c, int(a.ItemPct))
		}
	case rules.ClearCombat:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		return rules.Outcome{TurnOver: r.S.ClearCombatInfo(c)}, true, nil
	case rules.Configure:
		st, ok := a.With.(Settings)
		if !ok {
			return rules.Outcome{}, true, errBadValue
		}
		r.S.Settings = st
		return rules.Outcome{}, true, nil
	case rules.CastingTaken:
		c := r.chars[a.Who]
		if c == nil || c.Combat == nil {
			return rules.Outcome{}, true, errNoChar
		}
		c.Combat.Spell = 0
		return rules.Outcome{}, true, nil
	default:
		return rules.Outcome{}, false, nil
	}
	return rules.Outcome{}, true, nil
}

// factFight answers the Facts of the fight topic.
func (r *Rules) factFight(id rules.CharID, f rules.Fact) (rules.Value, bool, error) {
	s := r.S
	switch f := f.(type) {
	case rules.GroundFact:
		return r.ground(), true, nil
	case rules.UnitsFact:
		return r.units(), true, nil
	case rules.FlightFact:
		return rules.Flight{Kind: flight(f.Spell, f.Item)}, true, nil
	case rules.BeastKindFact:
		return rules.BeastKind{Name: BeastKind(f.Monster)}, true, nil
	case rules.UnitCharFact:
		return rules.UnitChar{Who: r.ID(r.S.Combatant(byte(f.Unit)))}, true, nil
	}
	c := r.chars[id]
	switch f.(type) {
	case rules.MayActFact, rules.RangeToFact, rules.SizeFact, rules.FightCardFact, rules.AimFact, rules.AimCheckFact,
		rules.AtFact, rules.StepFact, rules.UnitFact, rules.WeaponFact, rules.WeaponRulesFact, rules.TargetWhyFact, rules.SquaresFact, rules.ManeuversFact, rules.HitFact:
		if c == nil {
			return nil, true, errNoChar
		}
	default:
		return nil, false, nil
	}
	switch f := f.(type) {
	case rules.StepFact:
		who, g := s.GroundAt(c, f.Dir)
		return rules.StepInfo{Unit: int32(who), Ground: int32(g), Foe: who > 0 && s.Combatant(who).Side != c.Side}, true, nil
	case rules.AtFact:
		return rules.BoardPos{X: s.CombatX(c), Y: s.CombatY(c)}, true, nil
	case rules.UnitFact, rules.WeaponFact, rules.WeaponRulesFact, rules.TargetWhyFact, rules.SquaresFact, rules.ManeuversFact:
		v, err := r.fightFact(c, f)
		return v, true, err
	case rules.HitFact:
		b := r.chars[f.Target]
		if b == nil {
			return nil, true, errNoChar
		}
		return rules.Number(s.HitChance(c, b, s.Hand(c))), true, nil
	case rules.MayActFact:
		return rules.Yes(s.MayActAgainst(c, r.chars[f.Target])), true, nil
	case rules.RangeToFact:
		t := r.chars[f.Target]
		if t == nil {
			return nil, true, errNoChar
		}
		return rules.Number(s.RangeTo(c, t)), true, nil
	case rules.SizeFact:
		return rules.Number(s.CombatSize(c)), true, nil
	case rules.FightCardFact:
		card := rules.FightCard{Auto: c.Auto != 0}
		if ci := c.Combat; ci != nil {
			card.Delay, card.Moves, card.Casting = int(int8(ci.Delay)), int(ci.Moves), int(ci.Spell)
			card.Ranged, card.Attacked = ci.Ranged, ci.Attacked != 0
		}
		return card, true, nil
	case rules.AimFact:
		return r.aimOffer(c, f), true, nil
	case rules.AimCheckFact:
		list := make([]NearEntry, 0, len(f.Near))
		for _, n := range f.Near {
			list = append(list, NearEntry{Index: byte(s.CombatIndex(r.chars[n]))})
		}
		ground := 0
		if f.Ground {
			ground = 1
		}
		cand, why := s.AimCheck(c, s.Wielding(c, f.Ranged), list, f.At.X, f.At.Y, f.Range, f.ShowRange, ground)
		return rules.AimCheck{Target: r.ID(cand), Why: why, Occupied: s.Occ[f.At.Y][f.At.X] > 0}, true, nil
	}
	return nil, true, errNoFact
}

// fightCleared = por combatEnd's rule part: loadouts back, fight state,
// clouds and combat map reset (core.FightCleared).
func (r *Rules) fightCleared() {
	s := r.S
	s.RestoreLoadouts()
	s.Fight = FightState{}
	s.Clouds = nil
	s.Cmap = nil
}

// detectLoot = por LootDetect's cast (core.DetectLoot).
func (r *Rules) detectLoot() error {
	s := r.S
	n := s.LootDetect()
	if n == 0 {
		return rules.ErrNotNow
	}
	quit := false
	s.CastSpell(SpellID(n), 0, 0, CastReply{Fight: s.Fight.On}, &quit)
	return nil
}

// fightFact answers the turn view's facts of c (core.UnitFact ...).
func (r *Rules) fightFact(c *Character, f rules.Fact) (rules.Value, error) {
	s := r.S
	w := c.Slots[SlotWeapon]
	switch f := f.(type) {
	case rules.WeaponFact:
		if f.Ranged {
			w = s.rangedItem(c)
		}
	case rules.WeaponRulesFact:
		if f.Ranged {
			w = s.rangedItem(c)
		}
	}
	switch f := f.(type) {
	case rules.UnitFact:
		return rules.Number(s.CombatIndex(c)), nil
	case rules.WeaponFact:
		if w == nil {
			return rules.Handles{}, nil
		}
		return rules.Handles{r.Handle(w)}, nil
	case rules.WeaponRulesFact:
		if w == nil {
			return rules.Lines(nil), nil
		}
		return rules.Lines(s.WeaponRules(w)), nil
	case rules.TargetWhyFact:
		t := r.chars[f.Target]
		if t == nil {
			return nil, errNoChar
		}
		return rules.Text(s.TargetWhy(c, t, s.Hand(c), s.WeaponRange(s.Hand(c)), 1)), nil
	case rules.ManeuversFact:
		var out rules.ManeuverCards
		for _, m := range s.Rules().Maneuvers(s, c) {
			out = append(out, rules.ManeuverCard{ID: rules.ManeuverID(m.ID), Name: m.Name, Tip: m.Tip, Aim: rules.ManeuverAim(m.Aim), Range: m.Range, Area: m.Area, Move: m.Move})
		}
		return out, nil
	case rules.SquaresFact:
		return rules.Squares(s.Squares(c, f.Move)), nil
	}
	return nil, errNoFact
}

// aimOffer = por playerAim's opening (ovr14:3558): the near list round c
// and the units the aim offers (while one can be taken: only those).
func (r *Rules) aimOffer(c *Character, f rules.AimFact) rules.AimOffer {
	s := r.S
	rng := byte(f.Range)
	with := s.Wielding(c, f.Ranged)
	if rng == 0xFF {
		if w := with.Weapon; w != nil {
			rng = s.Base(w.Type).Range - 1
		} else {
			rng = 1
		}
	}
	if rng == 0 || rng == 0xFF {
		rng = 1
	}
	s.BuildNear(s.CombatX(c), s.CombatY(c), 0x7F, 0xFF, s.CombatSize(c)) // ovr14:32E6
	list := slices.Clone(s.Near)
	out := rules.AimOffer{Range: int(rng)}
	for _, n := range list {
		out.Near = append(out.Near, r.ID(s.Combatant(n.Index)))
	}
	foe := func(t *Character) bool { return t != nil && t.Side != c.Side && s.MayActAgainst(c, t) }
	offer := func(t *Character) bool { return t != nil && !f.FoesOnly || foe(t) }
	hitOnly := false
	for _, n := range list {
		if t := s.Combatant(n.Index); offer(t) && s.CanTarget(c, t, with, int(rng), f.ShowRange) {
			hitOnly = true
		}
	}
	for _, n := range list {
		t := s.Combatant(n.Index)
		if offer(t) && (!hitOnly || s.CanTarget(c, t, with, int(rng), f.ShowRange)) {
			hit := -1
			if hitOnly && f.ShowRange != 0 && foe(t) {
				hit = s.HitChance(c, t, with)
			}
			out.Targets, out.Hits = append(out.Targets, r.ID(t)), append(out.Hits, hit)
		}
	}
	return out
}

// blessSkips: applySpellToSide's test for an ally in Bless's area: one with
// an enemy next to it is left out. The near and target lists are kept.
func (r *Rules) blessSkips(t *Character) bool {
	s := r.S
	near, targets := s.Near, s.Targets
	s.Near = nil
	defer func() { s.Near, s.Targets = near, targets }()
	return s.NearEnemies(t, 1) > 0
}

// ground is the fight's ground and the footprints.
func (r *Rules) ground() rules.Ground {
	s := r.S
	var g rules.Ground
	if s.Cmap != nil {
		for y := range g.Squares {
			for x := range g.Squares[y] {
				g.Squares[y][x] = s.GroundSquare(byte(s.Cmap.Ground[y][x]))
				g.Squares[y][x].Crown = y+1 < len(g.Squares) && s.T().Gen.Wild.crownOver(s.Cmap.Ground[y][x], s.Cmap.Ground[y+1][x])
			}
		}
	}
	for size := range g.Footprints {
		for part := 0; part <= 3; part++ {
			if dx, dy, ok := s.Footprint(size, part); ok {
				g.Footprints[size] = append(g.Footprints[size], rules.BoardPos{X: dx, Y: dy})
			}
		}
	}
	return g
}

// units are the fighters on the board (por battle's loop).
func (r *Rules) units() rules.Units {
	s := r.S
	var out rules.Units
	for i := 1; i <= s.Ncomb; i++ {
		p, c := s.Cpos[i], s.Combatant(byte(i))
		if p.Size == 0 || c == nil || c.Combat == nil {
			continue
		}
		u := rules.Unit{ID: r.ID(c), Number: i, At: rules.BoardPos{X: int(p.X), Y: int(p.Y)}, Size: int(p.Size), Facing: int(c.Combat.Facing),
			Party: IsParty(c), Helpless: s.IsHelpless(c), Active: c.Okay != 0, Icon: int(c.Order),
			Name: c.Name(), HP: int(c.HP), MaxHP: int(c.MaxHP), AC: 60 - int(c.AC[0]), Damage: DamageText(c)}
		for part := 0; part <= 3; part++ {
			if dx, dy, ok := s.Footprint(u.Size, part); ok {
				u.Squares = append(u.Squares, rules.BoardPos{X: u.At.X + dx, Y: u.At.Y + dy})
			}
		}
		out = append(out, u)
	}
	return out
}

// inFight: sit is a fight's.
func inFight(sit rules.Situation) bool { _, ok := sit.(rules.Fight); return ok }

// openTurn = what por combatMenu (ovr09:0666) runs before its offer: affect
// group 0xF on the Actor, and a maneuver left armed is disarmed.
func (r *Rules) openTurn(a rules.OpenTurn) error {
	c := r.chars[a.Actor]
	if c == nil {
		return nil // as Allowed: nobody to offer to
	}
	r.S.AffectGroup(c, 0xF)
	r.S.Fight.Armed = Armed{}
	return nil
}
