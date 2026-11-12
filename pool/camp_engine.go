package pool

import (
	"gobox/core"
	"gobox/ega"
	"strconv"
)

// restFor (extra) rests mins minutes or until stop says so; then follows
// (restOver), getting whether a monster interrupted.
func (e *Engine) restFor(mins int, untilHealed bool, then restThen) {
	e.must(core.SetRest{Minutes: int32(mins)})
	e.Rest(untilHealed, then)
}

// healParty (extra) is the later games' Fix: clerics cast their Cure Light
// Wounds on the most wounded; while that is not enough they memorize Cure
// Light Wounds in every first-level slot, rest to learn it and cast again.
// Then they rest to learn their planned spells (Plan) again, and natural
// healing (1 HP a day) does the rest.
// The rests follow one another (restThen); the last ends in restStudy.
func (e *Engine) healParty() {
	e.castCures()
	e.healCures()
}

// healNatural: natural healing does the rest.
func (e *Engine) healNatural() {
	if e.healMinutes() > 0 {
		e.restFor(e.healMinutes(), true, restThen{kind: thHealDone})
		return
	}
	e.restStudy(false)
}

// healPlanned: the clerics rest to learn their planned spells again.
func (e *Engine) healPlanned() {
	if e.healClerics() == 0 {
		e.healNatural()
		return
	}
	e.must(core.ForgetCures{})
	if most := e.partyStudy(); most > 0 {
		e.restFor(most, false, restThen{kind: thHealPlanned})
		return
	}
	e.healNatural()
}

// healCures: while the cures are not enough the clerics memorize them,
// rest to learn them and cast again.
func (e *Engine) healCures() {
	if !(e.healClerics() > 0 && e.healMinutes() > 0 && e.ruleset().Check(core.SpellsAllowed{}, nil).Yes) {
		e.healPlanned()
		return
	}
	e.must(core.MemorizeCures{})
	e.restFor(e.partyStudy(), false, restThen{kind: thHealCured, n: e.healMinutes()})
}

// restOver goes on after a rest: what follows it, given whether it was
// interrupted.
func (e *Engine) restOver(then restThen, interrupted bool) {
	switch then.kind {
	case thRested:
		e.rested(interrupted)
	case thHealDone:
		e.restStudy(interrupted)
	case thHealPlanned:
		if interrupted {
			e.restStudy(true)
			return
		}
		e.healNatural()
	case thHealCured:
		if interrupted {
			e.restStudy(true)
			return
		}
		e.castCures()
		if e.healMinutes() >= then.n {
			e.healPlanned() // the cures do not help (nobody could cast)
			return
		}
		e.healCures()
	}
}

// restStudy ends RestFix's healing: the planned spells are learned only
// when there are clerics.
func (e *Engine) restStudy(interrupted bool) {
	if most := e.partyStudy(); !interrupted && most > 0 {
		e.restFor(most, false, restThen{kind: thRested})
		return
	}
	e.rested(interrupted)
}

// flagDay is State's Flags.Day: in camp the clock's (isDay, what Rest
// Until rests by, so its label and its rest agree), else the sky's.
func (e *Engine) flagDay(inside bool) bool {
	if e.campUp() {
		return e.ruleset().Check(core.IsDay{}, nil).Yes
	}
	return e.ruleset().Check(core.SkyDay{Inside: inside}, nil).Yes
}

// dropMember = ovr16:18E5 (the player confirmed): member n leaves the
// party; the last one ends the game.
func (e *Engine) dropMember(n int) {
	if n < 0 || n >= e.partySize() {
		return
	}
	e.selectMember(n)
	c := e.ruleset().Sheet(e.selected())
	if e.partySize() == 1 {
		e.RemoveSelected(false, true)
		e.Exit()
	}
	if c.Status == core.StatusOkay {
		e.say(c.Name + " bids you farewell")
	} else {
		e.say(c.Name + " is dumped in a ditch")
	}
	e.RemoveSelected(false, true)
}

// CampMenu = ovr16:1E6F, remade: camp takes decisions until the party
// leaves (verbClose) or a rest is interrupted; then what follows (the
// caller's cont) goes on, camp.interrupted telling whether it was (the
// caller then runs the encounter).
func (e *Engine) CampMenu() {
	if !e.songMusic(songInn, 0) { // extra: the inn theme for camp
		e.Music(1) // Amiga 0x17FBC .. 0x1813C
	}
	e.campSavedFlag = e.worldWord(core.WordBeginFlag)
	e.setWorldWord(core.WordBeginFlag, 2)
	e.must(core.CampOpen{})
	e.DrawMainScreen()
	e.must(core.ClearStudy{})
	e.campInterrupted = false
	e.campSavedAnim = e.animInput
	e.animInput = true // the camp picture's fire burns while camp is up (the camp bar's anim, ovr06:101F)
	e.setWorldWord(core.WordCampOn, 1)
	e.campSeq = e.opens()
}

// inCamp: camp's own decisions are open (no rest going on, no view over it).
func (e *Engine) inCamp() bool {
	return e.onTop() == onCamp && e.resting == nil && !e.view.on && !e.snap.invOn && e.iconEd == nil
}

// leaveCamp closes camp; what follows it runs.
func (e *Engine) leaveCamp() {
	if !e.campOn() {
		return
	}
	e.setWorldWord(core.WordCampOn, 0)
	e.must(core.CampClose{}) // the plans hold: the next rest learns them
	e.setWorldWord(core.WordBeginFlag, e.campSavedFlag)
	e.animInput = e.campSavedAnim
	e.ClearMessageArea()
	e.clearPrompt()
	e.Music(0)
	e.ret()
}

// rested ends a rest in camp: an interruption closes camp.
func (e *Engine) rested(interrupted bool) {
	e.campInterrupted = interrupted
	if interrupted {
		e.leaveCamp()
	}
}

// LeaveCamp = camp's Exit.
func (e *Engine) LeaveCamp() error {
	if !e.inCamp() {
		return core.ErrNotNow
	}
	e.leaveCamp()
	return nil
}

// PlanSpell plans member n's (< 0: the selected one) spell to be
// memorized count times on the next rest.
func (e *Engine) PlanSpell(n, spell, count int) error {
	if !e.inCamp() {
		return core.ErrNotNow
	}
	e.setWorldWord(core.WordWho, e.sel) // the rules' actor is the one selected before n
	e.must(core.Plan{Who: e.actor(n), Spell: e.spellID(spell), Count: int32(count)})
	return nil
}

// Scribe has member n scribe spell from a scroll on the next rest (again: not).
func (e *Engine) Scribe(n, spell int) error {
	if !e.inCamp() {
		return core.ErrNotNow
	}
	e.setWorldWord(core.WordWho, e.sel) // the rules' actor is the one selected before n
	e.must(core.Scribe{Who: e.actor(n), Spell: e.spellID(spell)})
	return nil
}

// DropMember = camp's and the party menu's Drop (the player asked twice).
func (e *Engine) DropMember(n int) error {
	switch {
	case e.inCamp():
		e.dropMember(n)
	case e.inParty():
		c := e.ruleset().Sheet(e.actor(n))
		if c.Status != core.StatusOkay {
			e.say("You dump " + c.Name + " out back.")
		} else {
			e.say(c.Name + " bids you farewell.")
		}
		e.RemoveSelected(false, true) // ovr17:3291
	default:
		return core.ErrNotNow
	}
	return nil
}

// CampRest = camp's Rest: RestFix, RestWait or RestUntil. The rest goes
// on in RestStep, and ends in rested.
func (e *Engine) CampRest(kind int) error {
	if !e.inCamp() {
		return core.ErrNotNow
	}
	e.must(core.PlanParty{}) // extra: the planned spells are learned on any rest
	switch kind {
	case RestFix:
		if e.healMinutes() == 0 && e.partyStudy() == 0 {
			e.say("Nothing to fix")
			return nil
		}
		if e.healMinutes() == 0 { // healing ends with the planned spells only when there are clerics
			e.restStudy(false)
		} else {
			e.healParty()
		}
	case RestWait:
		e.restFor(60, false, restThen{kind: thRested})
	case RestUntil:
		until := 20 * 60
		if !e.ruleset().Check(core.IsDay{}, nil).Yes {
			until = 6 * 60
		}
		e.restFor(((until-e.worldWord(core.WordClockMinutes))%(24*60)+24*60)%(24*60), false, restThen{kind: thRested})
	default:
		return core.ErrNotNow
	}
	return nil
}

// showRest = ovr21:05F0 (unit): "Rest time dd:hh:mm" (the original: "Rest Time:"), the unit being
// changed in white.
func (e *Engine) showRest(unit int) {
	var col [7]int
	for i := range col {
		col[i] = 0xA
	}
	col[unit] = 0xF
	r := e.restUnit
	e.text(1, 17, ega.LightGreen, "Rest time")
	x := 0xB
	for i := 4; i >= 3; i-- {
		e.text(x+1, 17, col[i], twoDigits(r(i)))
		e.text(x+3, 17, ega.LightGreen, ":")
		x += 3
	}
	e.text(x+1, 17, col[2], twoDigits(10*r(2)+r(1)))
}

// Rest = ovr21:0C9A: rest in five-minute steps until the time is up, a key
// stops it, or a monster interrupts (game +0x5A4 steps between rolls,
// +0x5A6 percent chance); then restOver gets whether it was interrupted.
// The loop is the presenter's to pace (RestStep), a key asks to stop it
// (StopRest).
func (e *Engine) Rest(untilHealed bool, then restThen) {
	e.must(core.BeginRest{})
	e.clearPane()
	e.showRest(0)
	e.restLine, e.restShown = true, 0
	e.resting = &restRun{untilHealed: untilHealed, then: then}
}

// RestStep rests five minutes on (the presenter paces the rest, as the
// DOS loop ran flat out); the rest over, camp goes on.
func (e *Engine) RestStep() error {
	run := e.resting
	if run == nil {
		return core.ErrNotNow
	}
	s := e.restTick(run.untilHealed)
	if e.restShown++; e.restShown >= 5 { // the rest's time on the page
		e.showRest(0)
		e.restShown = 0
	}
	if s.healed {
		e.showRest(0)
		e.text(1, 19, ega.LightGreen, "The Whole Party Is Healed")
		e.ClearMessageArea()
	}
	if s.interrupted {
		e.ClearMessageArea()
		e.showRest(0)
		e.text(1, 19, ega.White, "The Party is rudely interrupted!")
	}
	if s.over {
		e.resting = nil
		e.clearPane()
		e.restLine = false
		e.must(core.SetRest{})
		e.restOver(run.then, s.interrupted)
	}
	return nil
}

// StopRest = keyHit's "Stop resting?" answered Yes (the presenter asks):
// the rest ends at its next step.
func (e *Engine) StopRest() error {
	if e.resting == nil {
		return core.ErrNotNow
	}
	e.must(core.StopRest{})
	return nil
}

const (
	thRested      restThenKind = iota // camp's rest: an interruption closes camp
	thHealDone                        // healParty's last rest: its end
	thHealPlanned                     // healParty: the planned spells are learned
	thHealCured                       // healParty: the cures are memorized and learned
)

// restThen is what follows a rest (it gets whether a monster interrupted).
type restThen struct {
	kind restThenKind
	n    int // thHealCured: the healing minutes before the cures
}

type restThenKind byte

// restRun is a rest going on: RestStep takes five minutes of it.
type restRun struct {
	untilHealed bool // it ends too when nobody is wounded (healNatural)
	then        restThen
}

// restStep is what five minutes of rest did.
type restStep struct {
	over        bool // the rest is over
	healed      bool // the whole party healed
	interrupted bool // a monster interrupted it: over too
}

// restTick is StepRest ovr21:0C9A's five minutes: over before the step
// (RestDone), else the rules' Elapse.
func (e *Engine) restTick(untilHealed bool) (s restStep) {
	if e.ruleset().Check(core.RestDone{UntilHealed: untilHealed}, nil).Yes {
		s.over = true
		return
	}
	for _, f := range e.ruleset().Elapse(e.partyIDs(), 5, nil).Effects {
		switch f.(type) {
		case core.PartyHealed:
			s.healed = true
		case core.RestInterrupted:
			s.over, s.interrupted = true, true
		}
	}
	return
}

// Rest choices (CampRest).
const (
	RestFix   = iota // until everyone is healed (healParty) and the planned spells are learned
	RestWait         // an hour
	RestUntil        // until day (at night) or night (by day)
)

// restUnit is unit i of the rest time left (core.RestUnit: 1 minutes, 2 tens, 3 hours, 4 days).
func (e *Engine) restUnit(i int) int {
	return int(e.ruleset().Check(core.RestUnit{Unit: int32(i)}, nil).Value)
}

// twoDigits = ovr21:0565.
func twoDigits(n int) string {
	s := strconv.Itoa(n)
	if n < 10 {
		s = "0" + s
	}
	return s[:2]
}

// castCures has the clerics cast Cure Light Wounds, the member selected the actor.
func (e *Engine) castCures() {
	e.setWorldWord(core.WordWho, e.sel)
	e.must(core.CastCures{})
}

func (e *Engine) healMinutes() int {
	return int(e.ruleset().Check(core.HealMinutes{}, nil).Value)
}

func (e *Engine) partyStudy() int {
	return int(e.ruleset().Check(core.PartyStudy{}, nil).Value)
}

func (e *Engine) healClerics() int {
	return int(e.ruleset().Check(core.HealClerics{}, nil).Value)
}
