package engine

// dasp plays one of event's files and reports whether the pack took the
// sound: false when it is off or has nothing for event.
func (e *Engine) dasp(event string) bool {
	s := &e.snd
	if event == "" || !e.Dasp.Load() || s.d == nil || s.typ == sndNone || s.dasp == nil {
		return false
	}
	return s.dasp.Play(event)
}

// daspDir is the pack's folder under the assets.
const daspDir = "dasp"

// speechDir is the adventure's voice-over, one .ogg per text key.
const speechDir = "speech"
