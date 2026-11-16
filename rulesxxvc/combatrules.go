package rulesxxvc

import (
	"cmp"
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// combatState is types.pas TCombatState's rule part, per fight.
type combatState struct {
	initiative, tieBreaker int
	attacks                [2]int // phase 1 (weapon or baseScale), phase 2 (scale)
	movement               int
	guarding               bool

	// The AI's (combatai.pas): approach pattern and target last the fight,
	// the rest one turn.
	approach     int
	target       rules.CharID
	specialDue   bool // Decide returned the MONSTER_SPECIAL Maneuver
	aiBegun      bool // CombatAiControl's turn start ran
	aiPrepare    bool // a CombatAiControl pass is due: PrepareAiAction mode 1, then AiMoveAttack
	aiSteps      int  // AiMoveAttack's rounds, per pass
	lastMoveDir  int  // GDVars.lastMoveDir, per pass
	stuck        int  // GDVars.stuckCount, per pass
	hasAttacked  bool // RollAttacks ran this round (BeginRound clears it)
	attacksDrawn int  // this round's attacks on it (NoteAttackDirection)
	fleeing      bool // isFleeing: CheckMorale's flight
}

// fighters is the fight's characters as partyHead lists them: the party, then
// the board's others (the monsters EclLoadMonster appended).
// fightParty is the party list the fight runs on: the one the call hands in (the unit tests'), else the ruleset's
// own (GDVars.partyHead, the monsters EclLoadMonster appended with it): the engine passes none.
func (r *Rules) fightParty(f rules.Fight) []rules.CharID {
	if len(f.Party) > 0 {
		return f.Party
	}
	return r.party
}

func (r *Rules) fighters(f rules.Fight) []rules.CharID {
	ids := slices.Clone(r.fightParty(f))
	for _, u := range f.Board.Units {
		if !slices.Contains(ids, u.ID) {
			ids = append(ids, u.ID)
		}
	}
	return slices.DeleteFunc(ids, func(id rules.CharID) bool { return r.chars[id] == nil })
}

func (r *Rules) BeginFight(f rules.Fight, _ *rules.RNG) rules.Outcome {
	var o rules.Outcome
	r.arena = arena{}
	f.Board = r.board(f.Board)
	r.layCorpses(&f.Board, r.fighters(f), func(id rules.CharID) (rules.BoardPos, bool) { return unitAt(f, id) }, &o) // PlaceAllCreatures
	r.fight = map[rules.CharID]*combatState{}
	r.round = 0
	r.tactics = [2]bool{}
	r.moveBonus = int(f.Terms.PartyMove)
	r.terms = f.Terms
	r.hpPercent = r.hostileHpPercent(f) // combatst.pas setup
	return o
}

// surprised is zone1's sideMask bit (side + 1) for c's side: WordSurprise.
func surprised(f rules.Fight, c *Character) bool {
	for _, s := range f.Terms.Surprised {
		if (s == rules.SideMonsters) == !c.team() {
			return true
		}
	}
	return false
}

// BeginRound rolls everyone's initiative (cbtacts.pas BeginRound, in
// partyHead order, a surprised side -6) and orders them as combat.pas
// OrderCreatures: higher initiative first, then the higher tie-breaker, else
// list order; initiative 0 too.
func (r *Rules) BeginRound(f rules.Fight, rng *rules.RNG) rules.Outcome {
	rng = r.stream(rng) // the engine passes none: the ruleset's own stream
	f.Board = r.board(f.Board)
	if r.fight == nil {
		r.fight = map[rules.CharID]*combatState{}
	}
	r.round++
	o := r.rollTactics(f, rng)
	type rolled struct {
		id        rules.CharID
		init, tie int
	}
	var all []rolled
	for _, id := range r.fighters(f) {
		c := r.chars[id]
		i, t := Initiative(rng, c.Enabled, c.Scores[rules.Dexterity], surprised(f, c))
		var ai combatState
		if was := r.fight[id]; was != nil {
			ai.approach, ai.target = was.approach, was.target
		}
		s := &combatState{initiative: i, tieBreaker: t, approach: ai.approach, target: ai.target,
			movement: r.RecalcMovement(c)} // BeginRound: in half steps
		r.recalcAttacksLeft(f, c, s)
		s.attacks[1] = attacksLeft(attackRate(c, c.Scale), f.Round)
		r.fight[id] = s
		all = append(all, rolled{id, i, t})
	}
	slices.SortStableFunc(all, func(a, b rolled) int {
		return cmp.Or(cmp.Compare(b.init, a.init), cmp.Compare(b.tie, a.tie))
	})
	r.initiative = r.initiative[:0]
	for _, x := range all {
		r.initiative = append(r.initiative, x.id)
	}
	return o
}

// battleTacticsSkill is C_Skill.BATTLE_TACTICS.
var battleTacticsSkill, _ = skillNo(xxvc.BattleTactics)

// rollTactics is cbtacts.pas RollTactics: per side the first enabled
// fighter in list order, okay, dodging or past its jetpack flight, not
// stunned, that makes a Battle Tactics roll (Normal) gives its side +1 to
// hit this round (RecalcStats).
// ponytail: the message says "the" for his/her; Character has no gender.
func (r *Rules) rollTactics(f rules.Fight, rng *rules.RNG) rules.Outcome {
	r.tactics = [2]bool{}
	var o rules.Outcome
	for _, id := range r.fighters(f) {
		c := r.chars[id]
		if !c.Enabled || c.Status != StOkay && c.Status != StDodging && c.Status != StUsedJetpack ||
			hasEffect(c, EffStunned) || r.tactics[side(c)] {
			continue
		}
		if g, _ := SkillCheck([]int8{skillValue(c, battleTacticsSkill)}, rules.Normal, rng); g >= rules.Made {
			r.tactics[side(c)] = true // read at c's side's next RecalcStats
			o.Effects = append(o.Effects, rules.Narrated{Text: c.Name + " makes the tactics roll."})
		}
	}
	return o
}

func (r *Rules) NextActor(rules.Fight, *rules.RNG) (rules.CharID, bool) {
	if len(r.initiative) == 0 {
		return rules.NoChar, false
	}
	id := r.initiative[0]
	r.initiative = r.initiative[1:]
	return id, true
}

// zeroGSkill is C_Skill.MANEUVER_IN_0G.
var zeroGSkill, _ = skillNo("maneuver-in-0g")

// BeginTurn is combat.pas CreatureAction up to the control: a dodge or a
// failed zero-g roll ends, an initiative of 0 skips the turn, then RollZeroG
// (cbtacts.pas): on a zero-g board an enabled creature with initiative above
// 1 rolls Maneuver in 0g, Normal; a failure is StFailedZeroG.
func (r *Rules) BeginTurn(f rules.Fight, id rules.CharID, rng *rules.RNG) rules.Turn {
	rng = r.stream(rng) // the engine passes none: the ruleset's own stream
	f.Board = r.board(f.Board)
	c, s := r.chars[id], r.fight[id]
	if c == nil || s == nil {
		return rules.Turn{Skipped: true}
	}
	s.guarding, s.aiBegun, s.aiPrepare, s.aiSteps = false, false, true, 0
	if c.Status == StDodging || c.Status == StFailedZeroG {
		c.Status = StOkay
	}
	// TURN_START: HELPLESS, SCARED and STUNNED end the actions (EffEndActions);
	// SCARED then goes. ACTION_START's STUNNED and SCARED add nothing after it.
	if helpless(c) || hasEffect(c, EffScared) || hasEffect(c, EffStunned) {
		endActions(s)
	}
	removeEffect(c, s, EffScared)
	if s.initiative <= 0 {
		return rules.Turn{Skipped: true}
	}
	if s.initiative == 20 {
		s.initiative = 19
	}
	r.RecalcStats(c)
	t := rules.Turn{Computer: c.Side != rules.SideParty}
	if f.Board.ZeroG && c.Enabled && s.initiative > 1 {
		g, _ := SkillCheck([]int8{skillValue(c, zeroGSkill)}, rules.Normal, rng)
		text := c.Name + " made Zero-G skill roll"
		if g < rules.Made {
			c.Status = StFailedZeroG
			s.movement >>= 2
			text = c.Name + " failed Zero-G skill roll"
		}
		t.Effects = append(t.Effects, rules.Narrated{Text: text})
	}
	return t
}

func (r *Rules) EndTurn(rules.Fight, rules.CharID, *rules.RNG) rules.Outcome { return rules.Outcome{} }

// EndRound is combat.pas EndOfRound's rule part: a sprint or a used jetpack
// ends, the clouds count down. The clock's minute is the engine's.
func (r *Rules) EndRound(f rules.Fight, _ *rules.RNG) rules.Outcome {
	var o rules.Outcome
	f.Board = r.board(f.Board)
	for _, id := range r.fighters(f) {
		if c := r.chars[id]; c.Status == StSprinting || c.Status == StUsedJetpack {
			c.Status = StOkay
		}
	}
	r.hpPercent = r.hostileHpPercent(f) // combat.pas EndRound
	r.countDownClouds(&f.Board, &o)
	return o
}

// Verdict is CombatMain's end test over CountTeams (enabled per side): no
// team left is lost (fled when one ran), no hostile left is won.
// ponytail: combatEndRound is the engine's round limit, not tested here.
func (r *Rules) Verdict(f rules.Fight) rules.Verdict {
	var team, hostile int
	fled := false
	for _, id := range r.fighters(f) {
		c := r.chars[id]
		if c.team() && c.Status == StFleeing {
			fled = true
		}
		if !c.Enabled {
			continue
		}
		if c.team() {
			team++
		} else {
			hostile++
		}
	}
	switch {
	case team == 0 && fled:
		return rules.PartyFled
	case team == 0:
		return rules.PartyLost
	case hostile == 0:
		return rules.PartyWon
	}
	return rules.FightOn
}

// EndFight is postcbt.pas DetermineCombatResults (not solo): the fight's own
// states end; on a win DetermineLootExp shares the downed hostiles' xpAward
// over the enabled members (AwardExperience) and takes their credits; then
// the fled come back, the dying go unconscious, the unconscious with HP wake.
// The beaten's items (not NotTreasure, unreadied) are loot unless the terms
// skip them, last beaten first (availableItems is pushed onto); with
// RevivesParty the dying and the unconscious without HP stand up with 1 HP.
func (r *Rules) EndFight(f rules.Fight, _ *rules.RNG) rules.FightResult {
	r.tactics[0] = false // combat.pas: tacticsSides[TEAM] := 0
	r.moveBonus = 0      // postcbt.pas: zone1 movementBonus := 0
	r.arena = arena{}
	ids := r.fighters(f)
	for _, id := range ids {
		if c := r.chars[id]; c.Status == StOkay || c.Status >= StDodging && c.Status <= StFailedZeroG {
			c.Status, c.Enabled = StOkay, true
		}
	}
	var res rules.FightResult
	r.pool = 0
	verdict := r.Verdict(f)
	if verdict == rules.PartyWon {
		var total int64
		var enabled []*Character
		for _, id := range ids {
			c := r.chars[id]
			if !c.team() && c.Status != StOkay && c.Status != StFleeing {
				total += c.XPAward
				res.Money.Copper += uint64(c.Coins)
				res.Beaten++
				if !f.Terms.NoItemLoot {
					for _, it := range c.Items {
						if t, ok := r.tables.Items[it.Ref.ID]; ok && !t.NotTreasure {
							res.Items = append([]rules.ItemHandle{r.lootHandle(it.Ref)}, res.Items...)
						}
					}
				}
			}
			if c.team() && c.Enabled {
				enabled = append(enabled, c)
			}
		}
		if len(enabled) > 0 {
			share := total / int64(len(enabled))
			res.Exp = int32(share)
			for _, c := range enabled {
				c.Exp += share
			}
		}
	}
	if verdict != rules.PartyLost && verdict != rules.PartyFled {
		for _, id := range r.fightParty(f) {
			c := r.chars[id]
			if c == nil {
				continue
			}
			switch {
			case c.Status == StFleeing, c.Status == StUnconscious && c.HP > 0:
				c.Status, c.Enabled = StOkay, true
			case f.Terms.RevivesParty && (c.Status == StDying || c.Status == StUnconscious):
				c.Status, c.Enabled, c.HP = StOkay, true, 1
			case c.Status == StDying:
				c.Status = StUnconscious
			}
		}
	}
	r.pool = rules.Copper(res.Money.Copper) // DetermineLootExp's poolCredits
	if verdict != rules.PartyLost {         // the monsters leave the party list and the ruleset's characters (CharGen.RemoveCreature, old engine 1F6B leaveField)
		r.party = slices.DeleteFunc(r.party, func(id rules.CharID) bool {
			c := r.chars[id]
			gone := c != nil && !c.team()
			if gone {
				delete(r.chars, id)
			}
			return gone
		})
	}
	r.fight = nil
	return res
}

// lootHandle names a loot item by handle (FightResult.Items); r.loot keeps
// what each handle is.
func (r *Rules) lootHandle(ref rules.ItemRef) rules.ItemHandle {
	if r.loot == nil {
		r.loot = map[rules.ItemHandle]rules.ItemRef{}
	}
	h := rules.ItemHandle(len(r.loot) + 1)
	r.loot[h] = ref
	return h
}
