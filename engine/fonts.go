package engine

import (
	"time"
)

// Note is a short message for the window to show over the picture (the
// font name after Ctrl-F) and when it was set.
func (e *Engine) Note() (string, time.Time) {
	if e.restLine && time.Since(e.ext.noteAt) > 2*time.Second { // extra: the rest's time left, live in the flash's place
		r := e.restUnit
		return "Rest time " + twoDigits(r(4)) + ":" + twoDigits(r(3)) + ":" + twoDigits(10*r(2)+r(1)), time.Now()
	}
	return e.ext.note, e.ext.noteAt
}
