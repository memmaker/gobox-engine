package engine

import (
	"gobox/core"
	"gobox/game"
	"gobox/game/play"
	"image"
	"gobox/rules1e"
)

// present applies the effects recorded and not yet shown, then fx, in order.
func (e *Engine) present(fx ...core.Effect) {
	pending := append(e.takePending(), fx...)
	for _, f := range pending {
		e.apply(f)
	}
}

// emit queues an effect the Engine makes (a door's sound) after those the
// rules queued, for present to show (design W3: not the rules' Emit).
func (e *Engine) emit(f core.Effect) { e.fx = append(e.takePending(), f) }

// takePending takes the queued effects: the Engine's, with the rules' after them.
// ponytail: takes the rules' Pending until every rule call returns an Outcome (then core.Pending or nothing).
func (e *Engine) takePending() []core.Effect {
	fx := append(e.fx, e.ruleset().(core.Pending).TakePending()...)
	e.fx = nil
	return fx
}

// apply shows one effect.
func (e *Engine) apply(f core.Effect) {
	switch f := f.(type) {
	case core.Narrated: // in a fight: the battle log (Log) and the window's note
		if f.More && len(e.ext.battleLog) > 0 {
			l := &e.ext.battleLog[len(e.ext.battleLog)-1]
			*l = core.LogJoin(*l, f.Text)
			return
		}
		e.ext.battleLog = append(e.ext.battleLog, f.Text)
		if f.More {
			return
		}
		e.say(f.Text)
	case core.Turned: // the presenter turns the icon (the original's combatTurn)
		e.event(play.Turned{Unit: int(f.Unit), Facing: int(f.Facing)})
	case core.Casting:
		e.ext.panel = e.unitChar(f.From) // the side pane shows the caster
		e.event(play.CastSpell{From: int(f.From), Spell: e.spellNumber(f.Spell), Ats: points(f.Ats)})
	case core.Shot:
		frames, delay := e.missileArt(f)
		ev := play.Shot{From: int(f.From), To: point(f.To), StepMs: delay} // the original's pace: 10 pointed, 20 tumbling, 50 spinning, 30 a spell
		if frames > 0 && e.spr != nil {                                    // the original's art choice, the frame hitColours put in the sprite
			ev.Art = gamePicture(e.picPix(e.spr, 0, false, -1))
		}
		e.event(ev)
	case core.Sounded: // in a fight: when the presenter plays the events up to here; else now
		if e.inFight() {
			e.event(play.Sounded{Sound: string(f.Cue)})
			return
		}
		e.dasp(string(f.Cue))
	case core.Walked:
		w := play.Walked{Unit: int(f.Unit)}
		for _, p := range f.Path {
			w.Path = append(w.Path, game.Step{At: point(p)})
		}
		e.event(w)
	case core.Attacked:
		s := &e.snap // the snapshot's last weapon attack
		s.strike = Strike{Seq: s.strike.Seq + 1, Hit: f.Hit, Damage: int(f.Damage)}
		s.striker = [2]core.CharID{e.unitChar(f.From), e.unitChar(f.To)}
		e.ext.panel = s.striker[0] // the side pane shows the attacker, as the original's
		e.event(play.Attacked{From: int(f.From), To: int(f.To), Hit: f.Hit, Damage: int(f.Damage)})
		e.hitTo = 0
		if f.Hit {
			e.hitTo = int(f.To)
		}
	case core.Struck:
		if int(f.Unit) == e.hitTo {
			e.hitTo = 0
			return
		}
		e.event(play.Struck{Unit: int(f.Unit), Damage: int(f.Damage)})
	case core.Fell: // with the body ovr32:0E11 leaves for a fight the blow ends (class 55)
		ev := play.Fell{Unit: int(f.Unit)}
		b := e.ruleset().Check(core.UnitBody{Unit: f.Unit}, nil)
		if x, y := int(b.Low), int(b.High); !e.inView(x, y) {
			e.recentre(x, y, 3) // ovr32:0E11: the fall brought into view
		}
		if t := int(b.Value); b.Yes && t/0x80 < len(e.sqrpaci) && e.sqrpaci[t/0x80] != nil {
			ev.Body = gamePicture(e.picPix(e.sqrpaci[t/0x80], t%0x80, false, int(groundColour(e.art(e.sqrpaci[0], 0)))))
		}
		e.event(ev)

	default:
		panic("present: an effect with no presentation")
	}
}

// missileArt puts what flies in the sprite (DS:6E36) and gives its frames
// and pace: a weapon as attackAnimation (ovr14:26B0) chose, a spell or power
// as its caller (hitColours slot, then flyMissile's frames and delay).
func (e *Engine) missileArt(f core.Shot) (frames, delay int) {
	switch f.Spell {
	case "":
	case rules1e.SpellStoneGaze, rules1e.SpellCharmGaze:
		e.hitColours(0x12)
		return 4, 0x2D
	case rules1e.SpellSpitAcid:
		e.hitColours(0x17)
		return 1, 0x1E
	case rules1e.SpellBreathWeapon, rules1e.SpellLightningBolt, rules1e.SpellJavelinOfLightning:
		e.hitColours(0x13)
		return 4, 0x32
	default:
		e.hitColours(0x12)
		return 4, 0x1E
	}
	dir := int(e.ruleset().Check(core.ArcDir{Start: f.Start, To: f.To}, nil).Value)
	typ := f.Item // an item with no flight art flies as the default
	frames, delay, base := 1, 0xA, 0xD
	switch {
	case typ == rules1e.ItemIDDart || typ == rules1e.ItemIDJavelin || typ == rules1e.ItemIDQuarrel || typ == rules1e.ItemIDSpear || typ == rules1e.ItemIDArrow:
		switch {
		case dir&1 == 1 && (dir == 3 || dir == 5):
			e.iconToSprite(base+1, 1, 0, dir == 5)
		case dir&1 == 1:
			e.iconToSprite(base+1, 0, 0, dir == 7)
		default:
			e.iconToSprite(base+dir%4, dir>>2, 0, false)
		}
	case typ == rules1e.ItemIDHandAxe || typ == rules1e.ItemIDClub || typ == rules1e.ItemIDHammer:
		e.hitColours(base + 3)
		frames, delay = 4, 0x32
	case typ == rules1e.ItemIDHolyWater || typ == rules1e.ItemIDOil:
		e.hitColours(base + 4)
		frames, delay = 4, 0x32
	default:
		if typ == rules1e.ItemIDSling {
			base++
		}
		e.iconToSprite(base+7, 0, 0, false)
		e.iconToSprite(base+7, 1, 1, false)
		frames, delay = 2, 0x14
	}
	return frames, delay
}

// spellNumber is the spell number of a dnd spell id.
func (e *Engine) spellNumber(id core.SpellID) int {
	n := e.ruleset().Check(core.SpellNumber{Spell: id}, nil)
	if id == "" || !n.Yes {
		panic("present: no spell number for " + string(id))
	}
	return int(n.Value)
}

// point is a board square as the presenter's image.Point.
func point(p core.BoardPos) image.Point { return image.Pt(p.X, p.Y) }

// points are board squares as the presenter's.
func points(ps []core.BoardPos) []image.Point {
	var out []image.Point
	for _, p := range ps {
		out = append(out, point(p))
	}
	return out
}

// unitChar is who board unit u is (core.UnitCharFact; NoChar: none).
func (e *Engine) unitChar(u core.UnitID) core.CharID {
	return e.fact(core.NoChar, core.UnitCharFact{Unit: u}).(core.UnitChar).Who
}
