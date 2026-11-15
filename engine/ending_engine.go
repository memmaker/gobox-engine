package engine

import (
	"gobox/adventure"
	"gobox/game"
	"gobox/rules"
)

// endScene is the end's pictures with what gobox shows them by: the party's
// members (adventure.Layer.MinParty) and the message delay (adventure.Layer.MsgDelayed).
func (e *Engine) endScene(cm game.Common) game.Scene {
	sc := *e.title.end
	sc.Common = cm
	in := *sc.Intro
	in.Party, in.MsgDelay = e.memberCount(), int(e.msgDelay)
	sc.Intro = &in
	return sc
}

// ending plays the adventure's end (its manifest's End, SetEnd). With ends the run is over
// after it (pod's save-and-quit: game.Ended), else the script goes on (the finale of
// POR and COTAB: the game is won and play continues). With no End there is nothing to play.
// Reports whether it waits.
func (e *Engine) ending(ends bool) bool {
	e.ext.endRun = ends
	switch {
	case e.title.endAnim:
		e.sequence = adventure.SeqEnding
	case e.title.end != nil:
		e.ext.ending = true
	default:
		if ends {
			e.event(game.Ended{})
		}
		return false
	}
	e.wait(PendSequence)
	return true
}

// endDone: Continue ended the end; the run ends with it, or the script goes on.
func (e *Engine) endDone() error {
	e.sequence, e.ext.ending = "", false
	if e.ext.endRun {
		e.event(game.Ended{}) // the run ends: the script stays waiting
		return nil
	}
	e.endEffects()
	e.resume()
	return nil
}

// endEffects carries out what the end sequence leaves in the game (adventure.Sequence.After), through the
// ruleset, and says its closing note.
func (e *Engine) endEffects() {
	for _, f := range e.title.after {
		switch f {
		case adventure.EffectHealParty:
			e.do(rules.HealParty{})
		case adventure.EffectOpenHall:
			e.do(rules.OpenHallForAll{})
		}
	}
	if e.title.note != "" {
		e.say(e.title.note)
	}
}
