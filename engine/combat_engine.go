package engine

// The combat main loop and the player's command menus (ovr09).

import (
	"cmp"
	"fmt"
	"gobox/core"
	"image"
	"slices"
)

// combatEnd = ovr09:0000: free the clouds, the map and the sprite.
func (e *Engine) combatEnd() {
	e.as(e.sel)
	if _, err := e.ruleset().Resolve(core.Fight{}, core.FightCleared{}, nil); err != nil {
		panic(err)
	}
	e.cs = combatState{}
	e.snap.cmapOn = false
	e.cview = image.Point{}
	e.ccursor = combatCursor{}
	e.spr = nil
	e.combatViewEnter()
}

// takeTurn = ovr09:01F6 (char): false when the turn is skipped; the
// player's turn may leave screens up, the cursor's erase (endTurn) waits
// for them.
func (e *Engine) takeTurn(cID core.CharID) bool {
	e.as(e.sel) // the turn's actor
	t := e.ruleset().BeginTurn(core.Fight{}, cID, nil)
	if t.Skipped {
		return false // no cursor erase when the turn is skipped
	}
	e.selectChar(cID)
	e.combatRedrawChar(cID, 2, 1)
	e.ext.panel = cID
	if e.fightCard(cID).Delay > 0 {
		if t.Computer {
			e.as(e.sel)
			e.computerTurn(cID)
		} else {
			e.playerTurn(cID)
		}
	}
	return true
}

// turnVerb resolves a of c's turn through core.Rules; true when the turn is used.
func (e *Engine) turnVerb(cID core.CharID, a core.Action) bool {
	e.as(e.sel)
	out, err := e.ruleset().Resolve(core.Fight{Actor: cID}, a, nil)
	if err != nil {
		panic(fmt.Sprintf("turn %T: %v", a, err))
	}
	return out.TurnOver
}

// turnAsk is the question a of c's turn asks first (Outcome.Confirm), "" for none.
func (e *Engine) turnAsk(cID core.CharID, a core.Action) string {
	out, err := e.ruleset().Resolve(core.Fight{Actor: cID}, a, nil)
	if err != nil {
		panic(fmt.Sprintf("turn %T: %v", a, err))
	}
	return out.Confirm
}

// selectChar makes c the selected member.
func (e *Engine) selectChar(cID core.CharID) {
	i := slices.Index(e.partyIDs(), cID) // the rules' PartyIndex: past the end when not in it
	if i < 0 {
		i = e.partySize()
	}
	e.selectMember(i)
}

// playerTurn = ovr09:031A (char): a spell being cast goes off (aimed
// first), otherwise the combat menu until the turn is used. extra: the
// menu is the View's Turn (cs.turn); the player acts through TurnChoice.
func (e *Engine) playerTurn(cID core.CharID) {
	if !e.fact(cID, core.FactActive).(core.Yes) {
		e.clearCombat(cID)
		return
	}
	if s := e.fightCard(cID).Casting; s > 0 {
		e.resolve(core.CastingTaken{Who: cID})
		e.aimSpell(cID, s, castWhat{kind: cwInfo, c: cID, n: s})
		return
	}
	e.cs.turn = &turnState{e.turnChoices(cID)}
}

// turnChoice carries out decision a of c's turn.
func (e *Engine) turnChoice(cID core.CharID, a TurnAct, walks map[int]core.BoardPos) (done bool) {
	turn := e.snap.turn
	if turn == nil { // guard: no options up (an aim over the turn)
		return false
	}
	e.snap.turn = nil
	if !slices.Contains(turn.Can, a.Verb) && a.Verb != VerbWalk { // a walk is no bar button
		return false
	}
	switch a.Verb {
	case VerbWalk:
		done = e.turnVerb(cID, core.Move{To: core.BoardPos{X: a.X, Y: a.Y}})
	case VerbAttack:
		done = e.attackUnit(cID, a.N, walks)
	case VerbManeuver:
		done = e.maneuver(cID, a.N, walks)
	case VerbCast:
		if slices.ContainsFunc(turn.Spells, func(s Spell) bool { return s.ID == a.N }) {
			cast := castWhat{kind: cwTurn, c: cID, n: a.N}
			if e.fact(cID, core.SpellAimFact{Spell: e.spellID(a.N)}).(core.SpellAim).Now { // cast now: aimed first (castSpell only begins a longer one)
				e.aimSpell(cID, a.N, cast)
			} else {
				e.castRun(cast, nil)
			}
		}
	case VerbTurn:
		done = e.turnVerb(cID, core.TurnUndead{})
	case VerbQuick:
		e.turnVerb(cID, core.Quick{All: a.N == 1}) // extra: Shift-U, the whole party
		done = true
		e.as(e.sel)
		e.computerTurn(cID)
	case VerbGuard:
		done = e.turnVerb(cID, core.Guard{})
	case VerbBandage:
		done = e.turnVerb(cID, core.Bandage{Who: cID})
	case VerbWait: // Done's Delay
		done = e.turnVerb(cID, core.Delay{})
	case VerbEnd: // Done's Quit
		done = e.turnVerb(cID, core.EndTurn{})
	case VerbFlee: // moveAttack (ovr09:09DF): a step off the map, "Flee:" answered Yes
		if q := e.turnAsk(cID, core.Flee{}); q != "" {
			if !e.cs.sure {
				e.cs.ask = q
				break
			}
			e.later(resume{kind: rFlee, c: cID})
		}
	}
	return done
}

// turnChoices = ovr09:0666 (combatMenu)'s offer, published as the View's
// Turn with the squares a walk reaches (with their steps) and the foes
// verbAttack takes (in reach of the weapon, or a walk away: returned, by
// combat index, where the walk up to it ends): the rules' Allowed.
func (e *Engine) turnChoices(cID core.CharID) map[int]core.BoardPos {
	e.as(e.sel)
	e.resolve(core.OpenTurn{Actor: cID}) // the menu runs affect group 0xF and disarms first
	allowed := e.ruleset().Allowed(core.Fight{Actor: cID})
	t := &Turn{Unit: e.unit(cID), Moves: e.fightCard(cID).Moves >> 1}
	if w := e.fact(cID, core.WeaponFact{}).(core.Handles); len(w) > 0 {
		t.Weapon = e.invItem(cID, w[0])
		t.Weapon.Rules = e.fact(cID, core.WeaponRulesFact{}).(core.Lines)
	}
	for _, o := range allowed.Offers {
		switch a := o.Action.(type) {
		case core.Attack:
			t.Can = append(t.Can, VerbAttack)
		case core.Guard:
			t.Can = append(t.Can, VerbGuard)
		case core.UseItem:
			t.Can = append(t.Can, VerbItems)
		case core.Cast:
			if a.Spell == "" {
				t.Can = append(t.Can, VerbCast)
			} else {
				t.Spells = append(t.Spells, Spell{ID: e.spellNumber(a.Spell), Name: o.Name})
			}
		case core.Maneuver:
			if !slices.Contains(t.Can, VerbManeuver) {
				t.Can = append(t.Can, VerbManeuver)
				t.Maneuvers = e.maneuvers(cID)
			}
		case core.TurnUndead:
			t.Can = append(t.Can, VerbTurn)
		case core.Bandage:
			t.Can = append(t.Can, VerbBandage)
		case core.Delay:
			t.Can = append(t.Can, VerbWait)
		case core.EndTurn:
			t.Can = append(t.Can, VerbEnd)
		case core.Quick:
			t.Can = append(t.Can, VerbQuick)
		case core.Flee:
			t.Can = append(t.Can, VerbFlee)
		}
	}
	steps := func(w core.Reach) []core.Step {
		out := make([]core.Step, len(w.Path))
		for k, at := range w.Path {
			out[k] = core.Step{At: at, Free: w.Free[k]}
		}
		return out
	}
	e.snap.reach, e.snap.paths = nil, map[image.Point][]core.Step{}
	for _, w := range allowed.Moves {
		e.snap.reach = append(e.snap.reach, image.Pt(w.At.X, w.At.Y))
		e.snap.paths[image.Pt(w.At.X, w.At.Y)] = steps(w)
	}
	walks := map[int]core.BoardPos{}
	e.snap.targets, e.snap.hits = nil, map[int]int{}
	for _, t := range allowed.Aimed {
		a, ok := t.Action.(core.Attack)
		if !ok {
			continue
		}
		i := e.unit(a.Target)
		e.snap.targets = append(e.snap.targets, i)
		e.snap.hits[i] = e.hitChance(cID, a.Target)
		if t.Walk != nil { // hovering the foe shows the walk up to it
			walks[i] = t.Walk.At
			for _, sq := range t.Squares {
				e.snap.paths[image.Pt(sq.X, sq.Y)] = steps(*t.Walk)
			}
		}
	}
	e.ext.panel = cID
	e.snap.turn = t
	return walks
}

// attackUnit (extra): verbAttack. Unit 0 aims first (combatMenu's Aim);
// a foe of the turn's targets is attacked, a melee fighter walking up to
// it first (Allowed's Walk, pathNextTo) when it stands out of reach.
func (e *Engine) attackUnit(cID core.CharID, n int, walks map[int]core.BoardPos) bool {
	if n == 0 {
		e.playerAim(cID, 0xFF, 1, 0, aimThen{kind: atTurn}) // as the original: Tab cycles allies too, no foe needed
		return false
	}
	t := e.unitID(n)
	if !slices.Contains(e.snap.targets, n) { // extra: the status row says why not
		why := "Not a target"
		if e.as(e.sel); t != core.NoChar && bool(e.fact(cID, core.MayActFact{Target: t}).(core.Yes)) {
			why = cmp.Or(string(e.fact(cID, core.TargetWhyFact{Target: t}).(core.Text)), "No path to it this turn")
		}
		e.refuse(why)
		return false
	}
	if to, ok := walks[n]; ok {
		if e.turnVerb(cID, core.Move{To: to}) {
			return true
		}
		if !e.fact(t, core.FactActive).(core.Yes) || e.fact(cID, core.RangeToFact{Target: t}).(core.Number) > 1 { // stopped short: the turn goes on
			return false
		}
	}
	found, _ := e.aimConfirm(cID, t, 1)
	e.later(resume{kind: rBool, ok: found})
	return false
}

// endRound = ovr09:0880 (var done): the clock, the dying, the end check.
func (e *Engine) endRound() (done bool) {
	e.as(e.sel) // the round's actor
	for _, f := range e.ruleset().EndRound(core.Fight{}, nil).Effects {
		e.BottomMessage(f.(core.Narrated).Text)
	}
	v := e.ruleset().Verdict(core.Fight{})
	e.combatCentre(e.cview.X+combatView.W/2, e.cview.Y+3, 0xFF, 8)
	done = v != core.FightOn
	return
}

// hitChance (extra) is a's chance in percent to hit b: the rules' HitFact.
// at is id's square on the fight's board.
func (e *Engine) at(id core.CharID) core.BoardPos { return e.fact(id, core.AtFact{}).(core.BoardPos) }

// unit is id's unit number on the fight's board (0: none).
func (e *Engine) unit(id core.CharID) int { return int(e.fact(id, core.UnitFact{}).(core.Number)) }

func (e *Engine) hitChance(aID, bID core.CharID) int {
	v, err := e.ruleset().Fact(aID, core.HitFact{Target: bID})
	if err != nil {
		panic(err)
	}
	return int(v.(core.Number))
}

// combatUseHook = 00BA:2B88: back to the combat map: its palette, the map
// centred on the view (extra: no frame drawn: views draw their own).
func (e *Engine) combatUseHook() {
	e.combatViewLeave() // ovr32:00E3: the combat palette
	e.combatCentre(e.cview.X+combatView.W/2, e.cview.Y+3, 0xFF, 8)
}

// fightCard is id's turn record.
func (e *Engine) fightCard(id core.CharID) core.FightCard {
	return e.fact(id, core.FightCardFact{}).(core.FightCard)
}

// clearCombat clears id's turn record (ClearCombatInfo); its answer.
func (e *Engine) clearCombat(id core.CharID) bool {
	return e.resolve(core.ClearCombat{Who: id}).TurnOver
}

// unitID is who stands as unit n on the fight's board (NoChar: nobody).
func (e *Engine) unitID(n int) core.CharID {
	for _, u := range e.fact(core.NoChar, core.UnitsFact{}).(core.Units) {
		if u.Number == n {
			return u.ID
		}
	}
	return core.NoChar
}
