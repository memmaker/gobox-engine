package engine

import (
	"fmt"
	"gobox/game"
	"gobox/rules"
	"slices"
	"strconv"
)

// event records ev, after the effects before it (present).
func (e *Engine) event(ev game.Event) {
	e.present()
	e.events = append(e.events, ev)
}

// pushCombat = Combat (ovr09:007A): set up the battle, then rounds until
// one side is down (or the party stops fighting), a timed screen.
func (e *Engine) pushCombat() {
	e.setWorldWord(rules.WordBeginFlag, 5)
	e.present() // what the last fight left
	e.ext.battleLog, e.ext.panel = nil, rules.NoChar
	e.cs = combatState{}
	e.combatSetup()
	if f, ok := e.build().(game.Fighting); ok { // the board before the first turn: the step's fight events play from it
		e.event(game.Engaged{At: f})
	}
	if e.ruleset().Verdict(rules.Fight{}) != rules.FightOn {
		e.cs.phase = fightOver
	}
	e.wait(PendFight)
}

// CombatStep plays the fight on to the next turn taken (the presenter
// paces it); a player's turn then waits on TurnChoice. The fight over,
// the script goes on.
func (e *Engine) CombatStep() error {
	e.takeSettings()
	if !e.waiting(PendFight) || e.cs.actor != rules.NoChar {
		return rules.ErrNotNow
	}
	if e.stepFight() {
		e.resume()
	}
	e.present()
	return nil
}

// stepFight plays the fight on to the next turn taken; true when it is over.
func (e *Engine) stepFight() bool {
	for {
		switch e.cs.phase {
		case fightOver:
			return true
		case fightRound:
			e.as(e.sel) // the round's actor
			e.ruleset().BeginRound(rules.Fight{}, nil)
			e.cs.phase = fightTurn
		case fightTurn:
			id, ok := e.ruleset().NextActor(rules.Fight{}, nil)
			if !ok {
				e.cs.phase = fightEnd
				continue
			}
			m := e.ext.opened
			if !e.takeTurn(id) {
				e.ruleset().EndTurn(rules.Fight{}, id, nil)
				e.sidesCheck()
				continue
			}
			if e.cs.turn != nil || e.cs.aim != nil || e.ext.opened != m { // the player's turn: it ends when its turn and aim are done
				e.cs.actor = id
				return false
			}
			e.endTurn(id)
			return false
		case fightEnd:
			e.cs.phase = fightRound
			if e.endRound() {
				if e.ruleset().Verdict(rules.Fight{}) == rules.PartyWon { // ovr09:0971: the foes gone, "Continue Battle:" (Yes plays on)
					e.cs.phase = fightAsk
					e.askYesNo(yesNo("Continue Battle:", true), false, askDone{kind: dContinueFight}) // ovr09:099E asks through ovr26:10D1, Y/N only
					e.openAsk()
					return false
				}
				e.cs.phase = fightOver
				return true
			}
		case fightAsk: // the question is up
			return false
		}
	}
}

// endTurn = takeTurn's end: the cursor off; the last one down ends the round.
func (e *Engine) endTurn(cID rules.CharID) {
	e.ruleset().EndTurn(rules.Fight{}, cID, nil) // extra: the rule set's
	e.sidesCheck()
}

// sidesCheck: extra: the last one down ends the battle at once (the round ends).
func (e *Engine) sidesCheck() {
	if e.ruleset().Verdict(rules.Fight{}) != rules.FightOn {
		e.cs.phase = fightEnd
	}
}

// later hands the turn its result.
func (e *Engine) later(r resume) { e.cs.resume = r }

// resumed runs the result a choice left (later), if any.
func (e *Engine) resumed() bool {
	r := e.cs.resume
	e.cs.resume = resume{}
	switch r.kind {
	case rBool:
		return r.ok
	case rClear:
		return r.ok && e.clearCombat(r.c)
	case rFlee:
		return e.turnVerb(r.c, rules.Confirmed{Action: rules.Flee{}})
	}
	return false
}

// TurnChoice is decision a of the member whose turn it is (Battle.Turn):
// a choice of the turn's menu (Turn.Can, verbWalk, partyView), or while it
// aims (Turn.Aim) verbPoint, verbAim, verbClose. sure answers the question
// the same call asked (cs.ask, turnActs makes it the pending one).
func (e *Engine) TurnChoice(a TurnAct, sure bool) error {
	c := e.cs.actor
	if !e.waiting(PendFight) || c == rules.NoChar {
		return rules.ErrNotNow
	}
	e.cs.sure = sure
	defer func() { e.cs.sure = false }()
	switch t := e.cs.turn; {
	case e.cs.aim != nil:
		e.aimApply(a)
	case t != nil:
		if e.turnChoice(c, a, t.walks) || e.cs.aim == nil && e.top() == onWait && e.resumed() {
			e.turnDone()
		}
	}
	e.turnOn()
	return nil
}

// aimApply gives a to the aim; taken or given up, its result goes on
// (a further aim, the turn's later).
func (e *Engine) aimApply(a TurnAct) {
	h := e.cs.aim
	if !h.apply(a) {
		return
	}
	e.cs.aim = nil
	h.leave()
	if e.cs.aim == nil && e.cs.turn != nil && e.top() == onWait && e.resumed() {
		e.turnDone()
	}
}

// turnDone: the turn's menu is used.
func (e *Engine) turnDone() {
	e.cs.turn = nil
	e.snap.turn, e.snap.reach, e.snap.paths, e.snap.targets, e.snap.hits = nil, nil, nil, nil, nil
}

// turnOn goes on with the turn after a decision, or a screen it opened
// (the inventory, the sheet) going: its menu again, or its end.
func (e *Engine) turnOn() {
	c := e.cs.actor
	if c == rules.NoChar || e.cs.aim != nil || e.cs.choice != nil || e.top() != onWait {
		return
	}
	if t := e.cs.turn; t != nil {
		if e.cs.resume.kind != rNone && e.resumed() {
			e.turnDone()
		} else {
			t.walks = e.turnChoices(c)
			return
		}
	}
	e.cs.actor = rules.NoChar
	e.endTurn(c)
}

// aimSpell takes the aims c's spell will ask for (chooseSpellTargets with
// the player aiming) before cast casts it, which then takes them in order:
// the rules run straight through, in the same order and with the same
// dice as when they asked in the middle.
func (e *Engine) aimSpell(cID rules.CharID, spell int, cast castWhat) {
	aim := e.fact(cID, rules.SpellAimFact{Spell: e.spellID(spell)}).(rules.SpellAim)
	if aim.Kind == rules.SpellAimNone {
		e.castRun(cast, nil)
		return
	}
	sa := &spellAim{c: cID, spell: spell, cast: cast, kind: aim.Kind, aoe: max(1, aim.Area)}
	e.cs.sa = sa
	switch aim.Kind {
	case rules.SpellAimOne:
		e.aimOne(0)
	case rules.SpellAimArea:
		e.aimOne(1)
	default: // up to Targets different targets, as chooseSpellTargets' loop
		sa.left = aim.Targets
		e.aimNext()
	}
}

// aimOne aims once for the spell (allowGround: the square may be bare).
func (e *Engine) aimOne(allowGround int) {
	sa := e.cs.sa
	e.ext.aoe, e.ext.aoeSpell = sa.aoe, sa.spell // the area preview, as chooseSpellTargets shows it
	e.as(e.sel)
	aim := e.fact(sa.c, rules.SpellAimFact{Spell: e.spellID(sa.spell)}).(rules.SpellAim)
	e.ext.enemiesOnly = aim.FoesOnly
	e.playerAim(sa.c, aim.Range, 0, allowGround, aimThen{kind: atSpell})
}

// aimNext is the loop of the targets up to kind&3: the next aim, or the cast.
func (e *Engine) aimNext() {
	if e.cs.sa.left == 0 {
		e.aimsDone()
		return
	}
	e.aimOne(0)
}

// aimsDone casts the spell with the aims taken.
func (e *Engine) aimsDone() {
	sa := e.cs.sa
	e.cs.sa = nil
	e.castRun(sa.cast, sa.aims)
}

// aimSpellTaken goes on once one of the spell's aims is gone.
func (e *Engine) aimSpellTaken(found bool, t aimTarget) {
	sa := e.cs.sa
	e.ext.aoe, e.ext.aoeSpell, e.ext.enemiesOnly = 0, 0, false
	if !found && len(sa.aims) == 0 && sa.cast.kind != cwInfo { // extra: the first aim given up drops a cast chosen this turn (the menu's, an item's): nothing spent, the menu again; one begun rounds ago (cwInfo) goes off at nothing
		e.cs.sa = nil
		return
	}
	sa.aims = append(sa.aims, rules.Aim{Found: found, Target: t.C, At: t.At})
	switch {
	case sa.kind == rules.SpellAimOne, sa.kind == rules.SpellAimArea:
		e.aimsDone()
		return
	case !found:
		sa.left--
	case slices.Contains(sa.picked, t.C):
		e.say("Already been targeted") // the cast logs it as it goes over the aims
	default:
		sa.picked = append(sa.picked, t.C)
		sa.left--
	}
	e.aimNext()
}

// castRun is the cast an aim led to, through core.Rules Resolve(Cast/UseItem).
func (e *Engine) castRun(w castWhat, aims []rules.Aim) {
	if w.kind != cwItem && !w.chosen {
		if aim := e.fact(w.c, rules.SpellAimFact{Spell: e.spellID(w.n)}).(rules.SpellAim); aim.Choose != "" && (w.kind == cwInfo || aim.Now) {
			e.askCastChoice(w, aims, aim)
			return
		}
	}
	var a rules.Action
	switch w.kind {
	case cwInfo, cwTurn:
		a = rules.Cast{Caster: w.c, Spell: e.spellID(w.n), Aims: aims, Begun: w.kind == cwInfo, Choice: w.choice}
	case cwItem:
		if !e.fact(w.c, rules.FightCardFact{}).(rules.FightCard).Auto {
			e.combatUseHook() // 00BA:2B88: the player's use goes back to the combat map
		}
		u := rules.UseItem{User: w.c, Item: w.it, Aims: aims}
		if !w.own {
			u.Spell = e.spellID(w.n)
		}
		a = u
	}
	e.as(e.sel)
	out, err := e.ruleset().Resolve(rules.Fight{Actor: w.c}, a, nil)
	if err != nil {
		panic(fmt.Sprintf("cast %T: %v", a, err))
	}
	if w.kind != cwInfo {
		e.later(resume{kind: rBool, ok: out.TurnOver})
	}
}

// askCastChoice puts the spell's own question up (the pending one, which
// must be answered); its reply casts w (castChosen).
func (e *Engine) askCastChoice(w castWhat, aims []rules.Aim, aim rules.SpellAim) {
	e.cs.choice = &castChoice{w: w, aims: aims}
	pq := &playQuestion{q: game.Question{Text: aim.Choose, Replies: aim.Choices, NoEsc: true}}
	for i := range aim.Choices {
		pq.calls = append(pq.calls, pendingCall{kind: callCastChoice, choice: int32(i)})
	}
	e.ext.pq = pq
}

// castChosen casts the waiting spell with reply i, then the turn goes on
// as after the aim that led to it (aimApply, TurnChoice).
func (e *Engine) castChosen(i int32) {
	ch := e.cs.choice
	if ch == nil {
		return
	}
	e.cs.choice = nil
	w := ch.w
	w.chosen, w.choice = true, i
	e.castRun(w, ch.aims)
	if e.cs.aim == nil && e.cs.turn != nil && e.top() == onWait && e.resumed() {
		e.turnDone()
	}
	e.turnOn()
}

// aimFinished goes on with an aim's results: found, and the target t.
func (e *Engine) aimFinished(th aimThen, found bool, t aimTarget) {
	switch th.kind {
	case atTurn:
		e.later(resume{kind: rBool, ok: found})
	case atSpell:
		e.aimSpellTaken(found, t)
	case atNothing:
	default:
		e.maneuverAimed(th, found, t)
	}
}

// members is the party's cards (Party).
func (e *Engine) members() []Member {
	var out []Member
	sel := e.rosterSel() // a member pick shows its candidate
	for i, id := range e.partyIDs() {
		card := e.fact(id, rules.MemberCardFact{}).(rules.MemberCard)
		if !card.Party { // the monsters loaded for a fight sit on the same list: no cards for them
			continue
		}
		sh := e.ruleset().Sheet(id)
		out = append(out, Member{
			Name: sh.Name, HP: int(sh.HP), MaxHP: int(sh.MaxHP),
			AC: int(sh.AC), Status: e.statusName(id), Selected: i == sel, Colour: e.rosterColour(id),
			Class: card.Class, HeadPix: e.charHead(id), Portrait: card.Portrait, Powers: e.memberPowers(id),
		})
		m := &out[len(out)-1]
		m.Exp, m.LevelExp, m.NextExp = int(sh.Exp), int(sh.LevelExp), int(sh.NextExp)
		{
			ready, slots := card.Ready, card.Slots
			if slots > 0 {
				m.Spells = strconv.Itoa(ready) + "/" + strconv.Itoa(slots)
				if e.campOn() {
					m.Memo = m.Spells
				}
			}
		}
	}
	return out
}

// PlayEventSound plays an EvSound when the presenter reaches it, so a
// fight's sounds go with what they belong to (a blow with its lunge),
// not all at once when the rules ran.
// event is the cue the fight queued: the number itself (playSound's, a
// step), an original sound's pack event (daspFx), or one only the pack has
// (the rules' "mon_hit", "death" ...: the pack plays it, as out of a fight).
func (e *Engine) PlayEventSound(n int, event string) {
	if event == string(rules.CueStep) {
		e.playNow(sndStep, "")
		return
	}
	if k, err := strconv.Atoi(event); err == nil {
		e.playNow(k, "")
		return
	}
	for k, c := range daspFx {
		if string(c) == event {
			e.playNow(k, event)
			return
		}
	}
	e.dasp(event)
}

// playNow is playSound(n, event) now, not queued as a fight's event.
func (e *Engine) playNow(n int, event string) {
	e.soundNow = true
	e.playSound(n, event)
	e.soundNow = false
}

const (
	fightRound fightPhase = iota // the round begins
	fightTurn                    // the next one to act
	fightEnd                     // the round ends
	fightOver                    // the battle is over
	fightAsk                     // "Continue Battle:" is up (ovr09:0971)
)

const (
	rNone  resumeKind = iota
	rBool             // the turn is used when ok
	rClear            // when ok, and the turn ends (clearCombatInfo c)
	rFlee             // c steps off the map and flees
)

const (
	cwInfo castKind = iota // the spell a turn's combat info holds
	cwTurn                 // the turn menu's Cast
	cwItem                 // an item's spell (UseItem)
)

const (
	atTurn           aimThenKind = iota // the turn is used when something was found
	atSpell                             // one of a spell's aims (aimSpell)
	atManeuverMove                      // the square a maneuver moves to first
	atManeuverAttack                    // the attack a maneuver carries
	atArea                              // an area maneuver's square
	atTarget                            // a maneuver's target
	atNothing                           // the results are not wanted
)

// combatState is the fight's own state: where in the round it is and whose
// turn has screens up.
type combatState struct {
	phase  fightPhase   // what the next frame does
	actor  rules.CharID // the member whose turn is on (it ends when its turn and aim are done)
	turn   *turnState   // the player's turn menu waits (TurnChoice)
	aim    *aimRun      // the player aims (playerAim; TurnChoice's verbPoint, verbAim, verbClose)
	sure   bool         // the call is sure: the question it asked (ask) is answered Yes
	ask    string       // the yes/no question the turn call asked (Flee:, Attack Ally?): turnActs makes it the pending one
	resume resume       // what a choice left for the turn (later)
	sa     *spellAim    // the aims a spell takes before it is cast (aimSpell)
	choice *castChoice  // a spell's own question is up (askCastChoice): the cast and the turn wait for its reply
}

// castChoice is a cast waiting for the reply to its spell's question.
type castChoice struct {
	w    castWhat
	aims []rules.Aim
}

type fightPhase int

// resume is the result a choice hands the turn: whether the turn is used,
// known once the screens it pushed are gone (applyTurn), or at once when
// it pushed none.
type resume struct {
	kind resumeKind
	ok   bool
	c    rules.CharID
}

type resumeKind byte

// turnState is the player's turn menu: where the walk up to each foe ends.
type turnState struct{ walks map[int]rules.BoardPos }

// castWhat is the cast an aim leads to.
type castWhat struct {
	kind castKind
	c    rules.CharID
	n    int              // the spell (cwInfo, cwTurn) or the scroll's spell byte (cwItem)
	it   rules.ItemHandle // cwItem
	own  bool
	// chosen: the spell's own question (SpellAim.Choose) is answered, choice
	// the reply's index.
	chosen bool
	choice int32
}

type castKind byte

// spellAim is aimSpell's state while the player aims: the aims taken so far.
type spellAim struct {
	c      rules.CharID
	spell  int
	cast   castWhat
	kind   rules.SpellAimKind
	aoe    int
	aims   []rules.Aim
	left   int // targets still to aim (kinds up to 8)
	picked []rules.CharID
}

// aimThen is what the results of an aim go on with.
type aimThen struct {
	kind aimThenKind
	c    rules.CharID
	id   int            // the maneuver
	from rules.BoardPos // the square the maneuver began on
}

type aimThenKind byte
