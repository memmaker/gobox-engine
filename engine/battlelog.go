package engine

// The battle log (Battle.Log): a ledger of every change in a fight, from
// the rules' typed effects (logEffect), then every unit's hit points,
// status and spell effects against what the log last told of them
// (reconcile), so a change no effect names is logged too. The rules'
// own text (Narrated) goes to the window's note only. BATTLE-LOG.md.

import (
	"fmt"
	"slices"

	"gobox/rules"
)

// ledgerMark is a unit as the log last told of it.
type ledgerMark struct {
	hp      int32
	status  string
	effects []string
}

// ledger is the battle log and what it last told of each unit.
type ledger struct {
	lines []string
	marks map[rules.CharID]ledgerMark
	names map[rules.CharID]string // each unit's name in the log, fixed when it is first seen (logName)
	out   map[rules.CharID]bool   // the units the log told are down or gone
	told  map[rules.UnitID]int32  // the blows' damage logged with them (Attacked) whose Struck is still to come
}

// logStart opens the fight's log: every unit as it stands, nothing logged.
func (e *Engine) logStart() {
	e.ext.log = ledger{marks: map[rules.CharID]ledgerMark{}, names: map[rules.CharID]string{}, out: map[rules.CharID]bool{}, told: map[rules.UnitID]int32{}}
	e.reconcile(false)
}

// logLine adds a line to the log.
func (e *Engine) logLine(format string, a ...any) {
	e.ext.log.lines = append(e.ext.log.lines, fmt.Sprintf(format, a...))
}

// unitName is unit u's name in the log (logName).
func (e *Engine) unitName(u rules.UnitID) string { return e.logName(e.unitChar(u)) }

// logName is id's name in the log, as it was first seen (reconcile): with
// its unit number when another unit has the same name (three SWORDSMEN:
// "SWORDSMAN 2"), so a unit keeps its name after the others fall.
func (e *Engine) logName(id rules.CharID) string {
	if n, ok := e.ext.log.names[id]; ok {
		return n
	}
	return e.name(id)
}

// nameUnits fixes the log's name of every unit not named yet (logName).
func (e *Engine) nameUnits(units rules.Units) {
	for _, me := range units {
		if _, ok := e.ext.log.names[me.ID]; ok {
			continue
		}
		n := me.Name
		if slices.ContainsFunc(units, func(u rules.Unit) bool { return u.ID != me.ID && u.Name == me.Name }) {
			n = fmt.Sprintf("%s %d", me.Name, me.Number)
		}
		e.ext.log.names[me.ID] = n
	}
}

// logEffect logs f, a fight's effect; the harm it names is accounted on
// its unit's mark, so reconcile does not log it again.
func (e *Engine) logEffect(f rules.Effect) {
	l := &e.ext.log
	switch f := f.(type) {
	case rules.Attacked:
		with := ""
		if f.With != "" {
			with = " with " + f.With
		}
		switch {
		case !f.Hit:
			e.logLine("%s misses %s%s", e.unitName(f.From), e.unitName(f.To), with)
		default:
			e.logLine("%s hits %s%s for %d", e.unitName(f.From), e.unitName(f.To), with, f.Damage)
			l.told[f.To] += f.Damage
			e.harmed(f.To, f.Damage)
		}
	case rules.Struck:
		told := min(l.told[f.Unit], f.Damage) // a blow's own harm, logged with it (a rule set may report it on the blow alone)
		l.told[f.Unit] -= told
		if f.Damage -= told; f.Damage <= 0 {
			return
		}
		switch by, from := f.By, e.unitName(f.From); {
		case f.From != 0 && by != "":
			e.logLine("%s's %s deals %s %d damage", from, by, e.unitName(f.Unit), f.Damage)
		case by != "":
			e.logLine("%s deals %s %d damage", by, e.unitName(f.Unit), f.Damage)
		default:
			e.logLine("%s takes %d damage", e.unitName(f.Unit), f.Damage)
		}
		e.harmed(f.Unit, f.Damage)
	case rules.Used:
		e.logLine("%s uses %s", e.unitName(f.From), f.What)
	case rules.Casting:
		who := e.unitChar(f.From)
		e.logLine("%s casts %s", e.logName(who), e.fact(who, rules.SpellAimFact{Spell: f.Spell}).(rules.SpellAim).Name)
	case rules.Fell:
		id := e.unitChar(f.Unit)
		if f.Why != "" {
			e.logLine("%s %s", e.logName(id), f.Why)
		} else {
			e.logLine("%s goes down: %s", e.logName(id), e.statusName(id))
		}
		l.out[id] = true
		if m, ok := l.marks[id]; ok {
			m.status = e.statusName(id)
			l.marks[id] = m
		}
	case rules.Fled:
		e.logLine("%s flees", e.logName(f.Unit))
		l.out[f.Unit] = true
	}
}

// harmed accounts damage on u's mark: what reconcile expects of its hit
// points (never below 0).
func (e *Engine) harmed(u rules.UnitID, damage int32) {
	id := e.unitChar(u)
	if m, ok := e.ext.log.marks[id]; ok {
		m.hp = max(0, m.hp-damage)
		e.ext.log.marks[id] = m
	}
}

// reconcile logs (tell) every unit's change no effect told of: hit points,
// status, spell effects gained and ended, and marks them as they stand.
func (e *Engine) reconcile(tell bool) {
	l := &e.ext.log
	units := e.fact(rules.NoChar, rules.UnitsFact{}).(rules.Units)
	e.nameUnits(units)
	for id := range l.marks { // gone from the board with no word of it (a gaseous escape)
		if !slices.ContainsFunc(units, func(u rules.Unit) bool { return u.ID == id }) {
			if tell && !l.out[id] {
				e.logLine("%s leaves the fight: %s", e.logName(id), e.statusName(id))
			}
			delete(l.marks, id)
		}
	}
	for _, u := range units {
		sh := e.ruleset().Sheet(u.ID)
		sh.Name = e.logName(u.ID)
		now := ledgerMark{hp: sh.HP, status: e.statusName(u.ID), effects: slices.Clone(sh.Effects)}
		was, known := l.marks[u.ID]
		l.marks[u.ID] = now
		if !tell {
			continue
		}
		if !known {
			e.logLine("%s joins the fight", sh.Name)
			continue
		}
		switch d := now.hp - was.hp; {
		case d > 0:
			e.logLine("%s regains %d hp (%d)", sh.Name, d, now.hp)
		case d < 0:
			e.logLine("%s loses %d hp (%d)", sh.Name, -d, now.hp)
		}
		if now.status != was.status {
			e.logLine("%s is %s", sh.Name, now.status)
		}
		for _, x := range now.effects {
			if !slices.Contains(was.effects, x) {
				e.logLine("%s: %s", sh.Name, x)
			}
		}
		for _, x := range was.effects {
			if !slices.Contains(now.effects, x) {
				e.logLine("%s: %s ends", sh.Name, x)
			}
		}
	}
}
