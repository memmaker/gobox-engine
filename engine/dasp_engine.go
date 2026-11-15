package engine

import (
	"fmt"
	"os"
	"time"
)

// traceSounds (GOBOX_TRACE=1, as the front's trace) logs to stderr every
// sound the engine plays itself: the pack's event, the driver's effect
// number, the voice's recording.
var traceSounds = os.Getenv("GOBOX_TRACE") != ""

// traceSound logs one sound played, when traceSounds.
func traceSound(format string, args ...any) {
	if traceSounds {
		fmt.Fprintf(os.Stderr, "trace %s sound %s\n", time.Now().Format("15:04:05.000"), fmt.Sprintf(format, args...))
	}
}

// dasp plays one of event's files and reports whether the pack took the
// sound: false when it is off or has nothing for event.
func (e *Engine) dasp(event string) bool {
	s := &e.snd
	if e.unheard || event == "" || !e.Dasp.Load() || s.d == nil || s.typ == sndNone || s.dasp == nil {
		return false
	}
	played := s.dasp.Play(event)
	traceSound("pack event %q played=%v", event, played)
	return played
}

// daspDir is the pack's folder under the assets.
const daspDir = "dasp"

// speechDir is the adventure's voice-over, one .ogg per text key.
const speechDir = "speech"
