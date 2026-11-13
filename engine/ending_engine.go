package engine

import (
	"gobox/game"
	"gobox/game/play"
	"gobox/importer"
	"gobox/rules"
	"maps"
)

// defaultMsgDelay is the message delay an end animation's FrameMs is stated at (init.go).
const defaultMsgDelay = 1

// msgDelayTicks is the original's frame delay factor: a frame lasts its Delay times (message delay + 3).
const msgDelayTicks = 3

// endScene is the end's pictures as the party and the message delay show them: only as many party
// pictures as the party has members (game.Layer.MinParty), and the animations' frames paced by the
// message delay (game.Layer.MsgDelayed), as the original's finale.
func (e *Engine) endScene(cm play.Common) play.Scene {
	sc := *e.title.end
	sc.Common = cm
	party := e.worldWord(rules.WordPartyCount)
	sc.Steps = maps.Clone(sc.Steps)
	for n, st := range sc.Steps {
		var layers []game.Layer
		for _, l := range st.Layers {
			if l.MinParty > party {
				continue
			}
			if l.MsgDelayed {
				l.FrameMs = l.FrameMs * (int(e.msgDelay) + msgDelayTicks) / (defaultMsgDelay + msgDelayTicks)
			}
			l.MinParty, l.MsgDelayed = 0, false
			layers = append(layers, l)
		}
		st.Layers = layers
		sc.Steps[n] = st
	}
	return sc
}

// ending plays the adventure's end (its manifest's End, SetEnd). With ends the run is over
// after it (pod's save-and-quit: play.Ended), else the script goes on (the finale of
// POR and COTAB: the game is won and play continues). With no End there is nothing to play.
// Reports whether it waits.
func (e *Engine) ending(ends bool) bool {
	e.ext.endRun = ends
	switch {
	case e.hasSequence(seqEnding):
		e.sequence = seqEnding
	case e.title.end != nil:
		e.ext.ending = true
	default:
		if ends {
			e.event(play.Ended{})
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
		e.event(play.Ended{}) // the run ends: the script stays waiting
		return nil
	}
	e.endEffects()
	e.resume()
	return nil
}

// trainAll is the training hall's word once the game is won: it teaches every class (the original's 255).
const trainAll = 255

// endEffects carries out what the end sequence leaves in the game (importer.Sequence.After), through the
// ruleset, and says its closing note.
func (e *Engine) endEffects() {
	for _, f := range e.title.after {
		switch f {
		case importer.EffectHealParty:
			e.do(rules.HealParty{})
		case importer.EffectOpenHall:
			e.setWorldWord(rules.WordTraining, trainAll)
		}
	}
	if e.title.note != "" {
		e.say(e.title.note)
	}
}
