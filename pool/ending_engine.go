package pool

import "gobox/game/play"

// ending = pod's end (O043:0000 after the save): its end sequence plays, then the
// front ends the run (play.Ended); with no end sequence the run ends at once.
func (e *Engine) ending() {
	if e.hasSequence(seqEnding) {
		e.sequence = seqEnding
		e.wait(PendSequence)
		return
	}
	e.event(play.Ended{})
}

// sequenceDone: Continue ends the full-screen animation; the ending's ends the run.
func (e *Engine) sequenceDone() error {
	name := e.sequence
	e.sequence = ""
	if name == seqEnding {
		e.event(play.Ended{}) // the run ends: the script stays waiting
		return nil
	}
	e.resume()
	return nil
}
