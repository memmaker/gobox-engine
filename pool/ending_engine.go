package pool

import "gobox/game/play"

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
	e.resume()
	return nil
}
