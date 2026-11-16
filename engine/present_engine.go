package engine

import (
	"fmt"
	"gobox/game"
	"gobox/rules"
)

// present applies the effects recorded and not yet shown, then fx, in order.
func (e *Engine) present(fx ...rules.Effect) {
	pending := append(e.takePending(), fx...)
	for _, f := range pending {
		e.apply(f)
	}
}

// emit queues an effect the Engine makes (a door's sound) after those the
// rules queued, for present to show (design W3: not the rules' Emit).
func (e *Engine) emit(f rules.Effect) { e.fx = append(e.takePending(), f) }

// takePending takes the queued effects: the Engine's, with the rules' after them.
// ponytail: takes the rules' Pending until every rule call returns an Outcome (then core.Pending or nothing).
func (e *Engine) takePending() []rules.Effect {
	fx := e.fx
	e.fx = nil
	if p, ok := e.ruleset().(rules.Pending); ok { // a ruleset whose every call returns its Outcome (XXVc) queues nothing
		fx = append(fx, p.TakePending()...)
	}
	return fx
}

// apply shows one effect.
func (e *Engine) apply(f rules.Effect) {
	switch f := f.(type) {
	case rules.Narrated: // in a fight: the battle log (Log) and the window's note
		if f.More && len(e.ext.battleLog) > 0 {
			l := &e.ext.battleLog[len(e.ext.battleLog)-1]
			*l = rules.LogJoin(*l, f.Text)
			return
		}
		e.ext.battleLog = append(e.ext.battleLog, f.Text)
		if f.More {
			return
		}
		e.say(f.Text)
	case rules.Turned: // the presenter turns the icon (the original's combatTurn)
		e.event(game.Turned{Unit: int(f.Unit), Facing: int(f.Facing)})
	case rules.Casting:
		e.ext.panel = e.unitChar(f.From) // the side pane shows the caster
		e.event(game.CastSpell{From: int(f.From), Spell: e.spellNumber(f.Spell), Ats: f.Ats})
	case rules.Shot:
		frames, delay := e.missileArt(f)
		ev := game.Shot{From: int(f.From), To: f.To, StepMs: delay} // the original's pace: 10 pointed, 20 tumbling, 50 spinning, 30 a spell
		if frames > 0 && e.spr != nil {                             // the original's art choice, the frame hitColours put in the sprite
			ev.Art = gamePicture(e.picPix(e.spr, 0, false, nil))
		}
		e.event(ev)
	case rules.Sounded: // in a fight: when the presenter plays the events up to here; else now
		if e.unheard {
			return
		}
		if e.inFight() {
			e.event(game.Sounded{Sound: string(f.Cue)})
			return
		}
		e.dasp(string(f.Cue))
	case rules.Walked:
		w := game.Walked{Unit: int(f.Unit)}
		for _, p := range f.Path {
			w.Path = append(w.Path, game.Step{At: p})
		}
		e.event(w)
	case rules.Attacked:
		s := &e.snap // the snapshot's last weapon attack
		s.strike = Strike{Seq: s.strike.Seq + 1, Hit: f.Hit, Damage: int(f.Damage)}
		s.striker = [2]rules.CharID{e.unitChar(f.From), e.unitChar(f.To)}
		e.ext.panel = s.striker[0] // the side pane shows the attacker, as the original's
		e.event(game.Attacked{From: int(f.From), To: int(f.To), Hit: f.Hit, Damage: int(f.Damage)})
		e.hitTo = 0
		if f.Hit {
			e.hitTo = int(f.To)
		}
	case rules.Struck:
		if int(f.Unit) == e.hitTo {
			e.hitTo = 0
			return
		}
		e.event(game.Struck{Unit: int(f.Unit), Damage: int(f.Damage)})
	case rules.Fell: // with the body ovr32:0E11 leaves for a fight the blow ends (class 55)
		ev := game.Fell{Unit: int(f.Unit)}
		b := e.ruleset().Check(rules.UnitBody{Unit: f.Unit}, nil)
		if t := int(b.Value); b.Yes && t/0x80 < len(e.sqrpaci) && e.sqrpaci[t/0x80] != nil {
			ev.Body = gamePicture(e.picPix(e.sqrpaci[t/0x80], t%0x80, false, e.sqrpaci[0]))
		}
		e.event(ev)

	case rules.Fled: // the unit left the board: the ruleset's GroundFact/UnitsFact no longer hold it, read at each state build
	case rules.TerrainSet: // the board is the ruleset's, read as GroundFact at each state build: no event to play (a corpse's terrain)
	default:
		panic(fmt.Sprintf("present: an effect with no presentation: %T", f))
	}
}

// missileArt puts what flies in the sprite (DS:6E36) and gives its frames
// and pace: a weapon as attackAnimation (ovr14:26B0) chose, a spell or power
// as its caller (hitColours slot, then flyMissile's frames and delay).
func (e *Engine) missileArt(f rules.Shot) (frames, delay int) {
	if f.Spell != "" {
		switch e.flight(f) {
		case rules.FlightGaze:
			e.hitColours(0x12)
			return 4, 0x2D
		case rules.FlightSpit:
			e.hitColours(0x17)
			return 1, 0x1E
		case rules.FlightBolt:
			e.hitColours(0x13)
			return 4, 0x32
		default:
			e.hitColours(0x12)
			return 4, 0x1E
		}
	}
	dir := int(e.ruleset().Check(rules.ArcDir{Start: f.Start, To: f.To}, nil).Value)
	frames, delay, base := 1, 0xA, 0xD // an item with no flight art flies as the default
	kind := e.flight(f)
	switch kind {
	case rules.FlightPointed:
		switch {
		case dir&1 == 1 && (dir == 3 || dir == 5):
			e.iconToSprite(base+1, 1, 0, dir == 5)
		case dir&1 == 1:
			e.iconToSprite(base+1, 0, 0, dir == 7)
		default:
			e.iconToSprite(base+dir%4, dir>>2, 0, false)
		}
	case rules.FlightTumbling:
		e.hitColours(base + 3)
		frames, delay = 4, 0x32
	case rules.FlightFlask:
		e.hitColours(base + 4)
		frames, delay = 4, 0x32
	default:
		if kind == rules.FlightSlingStone {
			base++
		}
		e.iconToSprite(base+7, 0, 0, false)
		e.iconToSprite(base+7, 1, 1, false)
		frames, delay = 2, 0x14
	}
	return frames, delay
}

// flight is what the shot is in flight, as the ruleset says.
func (e *Engine) flight(f rules.Shot) rules.FlightKind {
	return e.fact(rules.NoChar, rules.FlightFact{Spell: f.Spell, Item: f.Item}).(rules.Flight).Kind
}

// spellNumber is the spell number of a dnd spell id.
func (e *Engine) spellNumber(id rules.SpellID) int {
	n := e.ruleset().Check(rules.SpellNumber{Spell: id}, nil)
	if id == "" || !n.Yes {
		panic("present: no spell number for " + string(id))
	}
	return int(n.Value)
}

// unitChar is who board unit u is (core.UnitCharFact; NoChar: none).
func (e *Engine) unitChar(u rules.UnitID) rules.CharID {
	return e.fact(rules.NoChar, rules.UnitCharFact{Unit: u}).(rules.UnitChar).Who
}
