package rules1e

import (
	"gobox/core"
	"gobox/core/goldbox"
	"slices"
)

// The fight: the board, aims, units, loot and the turn (core/fight.go).

// checkFight answers the Checks of the fight topic.
func (r *Rules) checkFight(k core.Check) (core.Answer, bool) {
	s := r.S
	switch k := k.(type) {
	case core.InFight:
		return core.Answer{Yes: r.S.Fight.On}, true
	case core.ArcDir:
		dir := 0
		for !s.InArc(k.Start.X, k.Start.Y, k.To.X, k.To.Y, dir) {
			dir++
		}
		return core.Answer{Value: int32(dir)}, true
	case core.UnitBody:
		c, p := s.Combatant(byte(k.Unit)), s.Cpos[k.Unit]
		return core.Answer{
			Yes:   c.Combat != nil && c.Combat.Extra == 0,
			Value: int32(s.GroundSquare(byte(s.grounds().Body)).Tile),
			Low:   int32(p.X), High: int32(p.Y),
		}, true
	case core.BoardUp:
		return core.Answer{Yes: r.S.Cmap != nil}, true
	case core.FightRound:
		return core.Answer{Yes: true, Value: int32(r.S.CombatRound)}, true
	default:
		return core.Answer{}, false
	}
}

// resolveFight answers the Actions of the fight topic.
func (r *Rules) resolveFight(sit core.Situation, a core.Action) (core.Outcome, bool, error) {
	s := r.S
	switch a := a.(type) {
	case core.OpenTurn:
		return core.Outcome{}, true, r.openTurn(a)
	case core.CombatSay:
		s.CombatSay(r.chars[a.Who], a.Text, true)
		return core.Outcome{}, true, nil
	case core.AutoTurn:
		o, err := r.autoTurn(sit, a)
		return o, true, err
	case core.Guard, core.Delay, core.EndTurn, core.TurnUndead, core.Quick, core.Move, core.Flee, core.Attack:
		o, err := r.turnVerb(sit, a)
		return o, true, err
	case core.Confirmed:
		switch a.Action.(type) {
		case core.Flee, core.Attack:
			o, err := r.turnVerb(sit, a)
			return o, true, err
		}
		return core.Outcome{}, false, nil
	case core.Bandage:
		if _, ok := sit.(core.Fight); ok {
			o, err := r.turnVerb(sit, a)
			return o, true, err
		}
		return core.Outcome{}, false, nil
	case core.DetectLoot:
		return core.Outcome{}, true, r.detectLoot()
	case core.FightCleared:
		r.fightCleared()
	case core.FightPlaced:
		r.fightPlaced()
	case core.TakeLoot:
		return core.Outcome{}, true, r.takeLoot(a)
	case core.BankLoot:
		s.MoneySound(s.Loot)
		s.Purse.Add(s.Loot, goldbox.GoldBoxRates)
		s.Loot = Purse{}
	case core.BuildGround:
		r.buildGround(a)
	case core.PlaceFight:
		r.combatBlocks()
		return core.Outcome{Effects: r.placeCombatants(a.Walls)}, true, nil
	case core.ClearLoot:
		s.ClearLoot() // DS:67F4
		s.Treasure = nil
	case core.ClearTreasure:
		s.Treasure = nil
	case core.DevTreasure:
		s.Treasure = s.devItems(ItemGauntlets)
	case core.AddTreasureBlock:
		b := r.TreasureBlock(int(a.Block))
		if len(b) == 0 {
			return core.Outcome{}, true, errNoBlock
		}
		b, err := s.gameItems(b) // the block is the game's
		if err != nil {
			return core.Outcome{}, true, err
		}
		s.AddTreasureBlock(b)
	case core.SetSpeed:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		s.SetSpeed(c)
	case core.SetLoot:
		s.SetLootCoins(a.Coins)
	case core.AddLoot:
		s.Loot.Add(a.Loot, goldbox.GoldBoxRates)
	case core.AddLibraryItems:
		return core.Outcome{}, true, s.AddLibraryItems(a.IDs, a.Known)
	case core.AddRandomTreasure:
		if err := s.AddRandomTreasure(int(a.N)); err != nil {
			return core.Outcome{}, true, err
		}
		c := r.chars[a.Who]
		for _, it := range s.Treasure {
			s.ItemName(c, it, false) // 00BA:0442
		}
	case core.Rob:
		robbed := r.party(a.Robbed)
		share := float64(a.Pct) / 100 * float64(len(robbed)) / float64(max(len(s.Party), 1))
		for _, n := range []*uint64{&s.Purse.Gold, &s.Purse.Copper, &s.Purse.Gems, &s.Purse.Jewelry} {
			*n -= uint64(float64(*n) * share)
		}
		for _, c := range robbed {
			s.ItemSave(c, int(a.ItemPct))
		}
	case core.ClearCombat:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		return core.Outcome{TurnOver: r.S.ClearCombatInfo(c)}, true, nil
	case core.Configure:
		st, ok := a.With.(Settings)
		if !ok {
			return core.Outcome{}, true, errBadValue
		}
		r.S.Settings = st
		return core.Outcome{}, true, nil
	case core.CastingTaken:
		c := r.chars[a.Who]
		if c == nil || c.Combat == nil {
			return core.Outcome{}, true, errNoChar
		}
		c.Combat.Spell = 0
		return core.Outcome{}, true, nil
	default:
		return core.Outcome{}, false, nil
	}
	return core.Outcome{}, true, nil
}

// factFight answers the Facts of the fight topic.
func (r *Rules) factFight(id core.CharID, f core.Fact) (core.Value, bool, error) {
	s := r.S
	switch f := f.(type) {
	case core.GroundFact:
		return r.ground(), true, nil
	case core.UnitsFact:
		return r.units(), true, nil
	case core.UnitCharFact:
		return core.UnitChar{Who: r.ID(r.S.Combatant(byte(f.Unit)))}, true, nil
	}
	c := r.chars[id]
	switch f.(type) {
	case core.MayActFact, core.RangeToFact, core.SizeFact, core.FightCardFact, core.AimFact, core.AimCheckFact,
		core.AtFact, core.StepFact, core.UnitFact, core.WeaponFact, core.WeaponRulesFact, core.TargetWhyFact, core.SquaresFact, core.ManeuversFact, core.HitFact:
		if c == nil {
			return nil, true, errNoChar
		}
	default:
		return nil, false, nil
	}
	switch f := f.(type) {
	case core.StepFact:
		who, g := s.GroundAt(c, f.Dir)
		return core.StepInfo{Unit: int32(who), Ground: int32(g), Foe: who > 0 && s.Combatant(who).Side != c.Side}, true, nil
	case core.AtFact:
		return core.BoardPos{X: s.CombatX(c), Y: s.CombatY(c)}, true, nil
	case core.UnitFact, core.WeaponFact, core.WeaponRulesFact, core.TargetWhyFact, core.SquaresFact, core.ManeuversFact:
		v, err := r.fightFact(c, f)
		return v, true, err
	case core.HitFact:
		b := r.chars[f.Target]
		if b == nil {
			return nil, true, errNoChar
		}
		return core.Number(s.HitChance(c, b)), true, nil
	case core.MayActFact:
		return core.Yes(s.MayActAgainst(c, r.chars[f.Target])), true, nil
	case core.RangeToFact:
		t := r.chars[f.Target]
		if t == nil {
			return nil, true, errNoChar
		}
		return core.Number(s.RangeTo(c, t)), true, nil
	case core.SizeFact:
		return core.Number(s.CombatSize(c)), true, nil
	case core.FightCardFact:
		card := core.FightCard{Auto: c.Auto != 0}
		if ci := c.Combat; ci != nil {
			card.Delay, card.Moves, card.Casting = int(int8(ci.Delay)), int(ci.Moves), int(ci.Spell)
		}
		return card, true, nil
	case core.AimFact:
		return r.aimOffer(c, f), true, nil
	case core.AimCheckFact:
		list := make([]NearEntry, 0, len(f.Near))
		for _, n := range f.Near {
			list = append(list, NearEntry{Index: byte(s.CombatIndex(r.chars[n]))})
		}
		ground := 0
		if f.Ground {
			ground = 1
		}
		cand, why := s.AimCheck(c, list, f.At.X, f.At.Y, f.Range, f.ShowRange, ground)
		return core.AimCheck{Target: r.ID(cand), Why: why, Occupied: s.Occ[f.At.Y][f.At.X] > 0}, true, nil
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
		return core.ErrNotNow
	}
	quit := false
	s.CastSpell(SpellID(n), 0, 0, CastReply{Fight: s.Fight.On}, &quit)
	return nil
}

// fightFact answers the turn view's facts of c (core.UnitFact ...).
func (r *Rules) fightFact(c *Character, f core.Fact) (core.Value, error) {
	s := r.S
	w := c.Slots[SlotWeapon]
	switch f := f.(type) {
	case core.UnitFact:
		return core.Number(s.CombatIndex(c)), nil
	case core.WeaponFact:
		if w == nil {
			return core.Handles{}, nil
		}
		return core.Handles{r.Handle(w)}, nil
	case core.WeaponRulesFact:
		if w == nil {
			return core.Lines(nil), nil
		}
		return core.Lines(s.WeaponRules(w)), nil
	case core.TargetWhyFact:
		t := r.chars[f.Target]
		if t == nil {
			return nil, errNoChar
		}
		return core.Text(s.TargetWhy(c, t, s.WeaponRange(c), 1)), nil
	case core.ManeuversFact:
		var out core.ManeuverCards
		for _, m := range s.Rules().Maneuvers(s, c) {
			out = append(out, core.ManeuverCard{ID: core.ManeuverID(m.ID), Name: m.Name, Tip: m.Tip, Aim: core.ManeuverAim(m.Aim), Range: m.Range, Area: m.Area, Move: m.Move})
		}
		return out, nil
	case core.SquaresFact:
		return core.Squares(s.Squares(c, f.Move)), nil
	}
	return nil, errNoFact
}

// aimOffer = por playerAim's opening (ovr14:3558): the near list round c
// and the units the aim offers (while one can be taken: only those).
func (r *Rules) aimOffer(c *Character, f core.AimFact) core.AimOffer {
	s := r.S
	rng := byte(f.Range)
	if rng == 0xFF {
		if w := c.Slots[SlotWeapon]; w != nil {
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
	out := core.AimOffer{Range: int(rng)}
	for _, n := range list {
		out.Near = append(out.Near, r.ID(s.Combatant(n.Index)))
	}
	foe := func(t *Character) bool { return t != nil && t.Side != c.Side && s.MayActAgainst(c, t) }
	offer := func(t *Character) bool { return t != nil && !f.FoesOnly || foe(t) }
	hitOnly := false
	for _, n := range list {
		if t := s.Combatant(n.Index); offer(t) && s.CanTarget(c, t, int(rng), f.ShowRange) {
			hitOnly = true
		}
	}
	for _, n := range list {
		t := s.Combatant(n.Index)
		if offer(t) && (!hitOnly || s.CanTarget(c, t, int(rng), f.ShowRange)) {
			hit := -1
			if hitOnly && f.ShowRange != 0 && foe(t) {
				hit = s.HitChance(c, t)
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
func (r *Rules) ground() core.Ground {
	s := r.S
	var g core.Ground
	if s.Cmap != nil {
		for y := range g.Squares {
			for x := range g.Squares[y] {
				g.Squares[y][x] = s.GroundSquare(byte(s.Cmap.Ground[y][x]))
			}
		}
	}
	for size := range g.Footprints {
		for part := 0; part <= 3; part++ {
			if dx, dy, ok := s.Footprint(size, part); ok {
				g.Footprints[size] = append(g.Footprints[size], core.BoardPos{X: dx, Y: dy})
			}
		}
	}
	return g
}

// units are the fighters on the board (por battle's loop).
func (r *Rules) units() core.Units {
	s := r.S
	var out core.Units
	for i := 1; i <= s.Ncomb; i++ {
		p, c := s.Cpos[i], s.Combatant(byte(i))
		if p.Size == 0 || c == nil || c.Combat == nil {
			continue
		}
		u := core.Unit{ID: r.ID(c), Number: i, At: core.BoardPos{X: int(p.X), Y: int(p.Y)}, Size: int(p.Size), Facing: int(c.Combat.Facing),
			Party: IsParty(c), Helpless: s.IsHelpless(c), Active: c.Okay != 0, Icon: int(c.Order),
			Name: c.Name(), HP: int(c.HP), MaxHP: int(c.MaxHP), AC: 60 - int(c.AC[0]), Damage: DamageText(c)}
		for part := 0; part <= 3; part++ {
			if dx, dy, ok := s.Footprint(u.Size, part); ok {
				u.Squares = append(u.Squares, core.BoardPos{X: u.At.X + dx, Y: u.At.Y + dy})
			}
		}
		out = append(out, u)
	}
	return out
}

// inFight: sit is a fight's.
func inFight(sit core.Situation) bool { _, ok := sit.(core.Fight); return ok }

// openTurn = what por combatMenu (ovr09:0666) runs before its offer: affect
// group 0xF on the Actor, and a maneuver left armed is disarmed.
func (r *Rules) openTurn(a core.OpenTurn) error {
	c := r.chars[a.Actor]
	if c == nil {
		return nil // as Allowed: nobody to offer to
	}
	r.S.AffectGroup(c, 0xF)
	r.S.Fight.Armed = Armed{}
	return nil
}
