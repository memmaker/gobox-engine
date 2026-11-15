package rules1e

import (
	"gobox/rules"
	"slices"
)

// The world and script side of core.Rules (fold step E-R6): the script's
// questions about the party.

// checkWorld answers the Checks por's script asks (eclops.go).
func (r *Rules) checkWorld(k rules.Check) rules.Answer {
	switch k := k.(type) {
	case rules.ClockIs, rules.MemberCount, rules.ActingFor, rules.PartyEffect:
		return r.sessionCheck(k)
	case rules.AreaTileIs:
		return rules.Answer{Value: int32(r.S.AreaTile(int(k.X), int(k.Y), k.Marks))}
	case rules.PartyStrength:
		return rules.Answer{Value: int32(r.partyStrength(k.Party))}
	case rules.Census:
		return r.census(k)
	case rules.HasClass:
		for _, c := range r.party(k.Party) {
			if slices.Contains(porClassesOf[k.Class], c.Class) {
				return rules.Answer{Yes: true, Who: r.ID(c)}
			}
		}
	case rules.HasCondition:
		aff, ok := conditionAffects[k.Condition]
		if !ok {
			panic("rules1e: no affect for condition " + string(k.Condition))
		}
		for _, id := range k.Whom {
			if c := r.party(rules.Party{id})[0]; r.S.FindAffect(c, aff) != nil { // opCheckParty 00BA:2199: any status
				return rules.Answer{Yes: true, Who: id}
			}
		}
	case rules.HasAffect:
		aff, err := r.S.T().affectID(byte(k.Affect))
		if err != nil {
			panic(err) // a script names an affect the importer's affectIds lacks: a data bug
		}
		for _, id := range k.Whom {
			if r.S.FindAffect(r.party(rules.Party{id})[0], aff) != nil { // 00BA:2199: any status
				return rules.Answer{Yes: true, Who: id}
			}
		}
	case rules.HasItem:
		for _, id := range k.Whom { // opFindItem ovr04:266A: any status, the pool too
			for _, it := range r.S.Carried(r.party(rules.Party{id})[0]) {
				if r.S.kind(it.Type) == k.Item {
					return rules.Answer{Yes: true, Who: id}
				}
			}
		}
	case rules.HasMemorized:
		k1 := slices.Index(SpellIDs[:], k.Spell)
		if k1 <= 0 {
			panic("rules1e: no spell number for " + string(k.Spell))
		}
		for _, c := range r.party(k.Party) {
			if c.Status == StatusOkay && SpellSlot(c, SpellID(k1)) != 0xFF { // FindCaster
				return rules.Answer{Yes: true, Who: r.ID(c)}
			}
		}
	case rules.RollDice:
		return rules.Answer{Value: int32(r.S.Dice(int(k.N), int(k.Sides)))}
	case rules.RollRandom:
		return rules.Answer{Value: int32(r.S.Random(int(k.N)))}
	case rules.IsDay:
		return rules.Answer{Yes: r.S.IsDay()}
	case rules.MapTerrain:
		return rules.Answer{Value: int32(r.S.MapTerrain(k.X, k.Y))}
	}
	return rules.Answer{}
}

// elapse is a camp rest's time passing in its five-minute steps (stepRest,
// RestStep ovr21:0C9A): core.PartyHealed when the whole party healed,
// core.RestInterrupted (and no more steps) when a monster interrupted.
// The party is the rules' marching order (e.Party), as the original's.
func (r *Rules) elapse(minutes int32) (o rules.Outcome) {
	for range minutes / 5 {
		s := r.S.stepRest()
		if s.Healed {
			o.Effects = append(o.Effects, rules.PartyHealed{})
		}
		if s.Interrupted {
			o.Effects = append(o.Effects, rules.RestInterrupted{})
			return o
		}
	}
	return o
}

// SetLootCoins is opTreasure's coins as the loot (DS:67F4): the seven
// kinds, the last two gems and jewelry (CoinsToPurse).
func (e *RuleState) SetLootCoins(coins [7]uint16) { e.Loot = CoinsToPurse(coins) }

// conditionAffects is the affect each Condition a script asks about is
// (HasCondition): por's CheckParty 19 (0x13).
// ponytail: the one scripts use; add conditions when a script asks.
var conditionAffects = map[rules.Condition]AffectType{CondFindTraps: affFindTraps}

// partyStrength = opPartyStrength: a tenth of HP + 5 x (AC
// better than 0) + 5 x (THAC0 better than 0) + 8 x mage level + 4 x cleric
// level per member, summed into a byte (it wraps, as the original's).
func (r *Rules) partyStrength(p rules.Party) byte {
	var total byte
	for _, c := range r.party(p) {
		ac, th := int(c.AC[0]), int(c.THAC0)
		if ac > 0x3C {
			ac -= 0x3C
		} else {
			ac = 0
		}
		if th > 0x27 {
			th -= 0x27
		} else {
			th = 0
		}
		v := int(c.HP) + 5*ac + 5*th + 8*int(c.Levels[ClsMage]) + 4*int(c.Levels[ClsCleric])
		total += byte(v / 10)
	}
	return total
}

// census = opCheckParty ovr04:1435's low, high and average of a byte fact
// over the party; an empty party leaves low 0xFF, high and average 0.
func (r *Rules) census(k rules.Census) rules.Answer {
	lo, hi, sum, n := byte(0xFF), byte(0), 0, 0
	for _, id := range k.Party {
		f, err := r.Fact(id, k.Fact)
		if err != nil {
			panic("rules1e: census of " + err.Error())
		}
		v := byte(f.(rules.Number))
		n++
		lo, hi, sum = min(lo, v), max(hi, v), sum+int(v)
	}
	a := rules.Answer{Low: int32(lo), High: int32(hi)}
	if n > 0 {
		a.Avg = int32(byte(sum / n))
	}
	return a
}

// party is p's characters; an id the rules do not hold is a bug.
func (r *Rules) party(p rules.Party) []*Character {
	cs := make([]*Character, len(p))
	for i, id := range p {
		if cs[i] = r.chars[id]; cs[i] == nil {
			panic(errNoChar)
		}
	}
	return cs
}

// porClassesOf is the record classes that have a class (HasClass).
// ponytail: ranger only (opPartySurprise); add classes when a script asks.
var porClassesOf = map[rules.Class][]ClassID{rules.ClassRanger: {ClsRanger, ClsClericRanger}}

// resolveWorld resolves the world's actions; ok: a was one of them.
func (r *Rules) resolveWorld(s rules.Situation, a rules.Action) (o rules.Outcome, ok bool, err error) {
	switch a := a.(type) {
	case rules.AdvanceClock:
		r.S.AdvanceClock(int(a.Count), int(a.Unit))
		return o, true, nil
	case rules.Fixture:
		return o, true, r.fixture(a)
	case rules.SetClock, rules.SetMemberCount, rules.ActFor, rules.EnterPhase, rules.OpenHallForAll:
		return o, true, r.sessionResolve(a)
	case rules.DevPlace:
		r.S.DevPlace(a.Melee)
		return o, true, nil
	case rules.DuelClone:
		return o, true, r.duelClone(a)
	case rules.ScriptDamage:
		return r.scriptDamage(a), true, nil
	case rules.ChangeClass:
		return o, true, r.changeClass(a)
	case rules.HoldInReserve:
		return o, true, r.holdInReserve(a)
	case rules.BashDoor:
		at, isDoor := s.(rules.AtDoor)
		if !isDoor {
			return o, true, ErrNotNow
		}
		c := r.chars[a.Who]
		if c == nil {
			return o, true, errNoChar
		}
		opened, spent := r.S.bashDoor(c, at.Door.State == rules.DoorMagic)
		if spent {
			o.Effects = append(o.Effects, rules.DoorTriesSpent{})
		}
		if opened {
			o.Effects = append(o.Effects, rules.DoorGaveWay{Who: a.Who})
		}
		return o, true, nil
	case rules.PickLock:
		if _, isDoor := s.(rules.AtDoor); !isDoor {
			return o, true, ErrNotNow
		}
		c := r.chars[a.Who]
		if c == nil {
			return o, true, errNoChar
		}
		if r.S.Dice(1, 100) <= int(c.Thief[1]) && c.Status == StatusOkay { // pickLock ovr15:04DF: the dice roll for the dead too
			o.Effects = append(o.Effects, rules.DoorGaveWay{Who: a.Who})
		}
		return o, true, nil
	}
	return o, false, nil
}

// bashDoor is one member's try of bashDoor ovr15:0294 at a door, magically
// locked or not: opened, or spent when the member is too weak for a magic one.
func (e *RuleState) bashDoor(c *Character, magic bool) (opened, spent bool) {
	if c.Status != StatusOkay { // fix, as the Pascal reference: the original let the dead and the sleeping push too
		return false, false
	}
	d6 := func(max int) bool { return e.Dice(1, 6) <= max }
	str, exc := c.Stats[StatStr], c.Str00
	if magic {
		switch {
		case str == 0x12:
			switch {
			case exc >= 0x5B && exc <= 0x63:
				return e.Dice(1, 6) == 1, false
			case exc == 0x64:
				return d6(2), false
			}
			return false, true
		case str == 0x13 || str == 0x14:
			return d6(3), false
		case str == 0x15 || str == 0x16:
			return d6(4), false
		case str == 0x17:
			return d6(5), false
		case str == 0x18:
			return e.Dice(1, 8) <= 7, false
		case str == 0x19:
			return true, false
		}
		return false, true
	}
	switch {
	case str >= 3 && str <= 7:
		return e.Dice(1, 6) == 1, false
	case str >= 8 && str <= 0xF:
		return d6(2), false
	case str >= 0x10 && str <= 0x11:
		return d6(3), false
	case str == 0x12:
		switch {
		case exc <= 0x32:
			return d6(3), false
		case exc <= 0x63:
			return d6(4), false
		case exc == 0x64:
			return d6(5), false
		}
	case str > 0x12: // fix, as the Pascal reference: the original's table stopped at 18/00, so a giant's strength failed
		return d6(5), false
	}
	return false, false
}

// duelClone = DuelSetup's opponent (story.go): Of's sheet, no affects,
// copies of its items in reverse order, on the monsters' side.
func (r *Rules) duelClone(a rules.DuelClone) error {
	of := r.chars[a.Of]
	if of == nil {
		return errNoChar
	}
	cl := &Character{Sheet: of.Sheet} // no affects, no items yet
	cl.Okay = 1
	cl.SetName(a.Name)
	cl.Auto = 1
	cl.Side = SideMonsters
	cl.Control = 0xB2
	cl.Order = byte(a.Icon)
	r.S.Party = append(r.S.Party, cl)
	for _, it := range of.Items {
		c := *it
		cl.Items = append([]*Item{&c}, cl.Items...)
	}
	return nil
}
