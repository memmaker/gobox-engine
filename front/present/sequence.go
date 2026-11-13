// sequence.go plays a full-screen animation (Sequence: VACE's
// .ACF/.BMA, gobox/dax/vace) and ends it (SequenceDone) when it runs out
// or on a key, click or pad button.

package present

import (
	"gobox/game"
	"time"

	"gobox/dax/vace"
)

// SequenceUI is the animation being played.
type SequenceUI struct {
	Name  string
	Anim  *vace.Anim // nil: its files did not decode (it ends at once)
	start time.Time
	done  bool // SequenceDone sent
	cue   int  // Anim.Cues fired so far
	// Cues are the music cues (Anim.Cues) that came due in the last step; a
	// "stop" ends it when the animation ends. The presenter plays them.
	Cues []vace.Cue
}

// SequenceStep plays sq (nil: none) this frame on in and returns the
// frame to show (nil: none); skip: a key, click or button came.
func SequenceStep(h RuleHost, in *SequenceUI, sq *game.Sequence, skip bool) *vace.VaceFrame {
	if sq == nil {
		*in = SequenceUI{}
		return nil
	}
	now := time.Now()
	if in.Name != sq.Name || in.start.IsZero() {
		a, _ := vace.Load(filesDir, sq.Anim)
		*in = SequenceUI{Name: sq.Name, Anim: a, start: now}
	}
	in.Cues = nil
	if in.done {
		return in.last()
	}
	t := now.Sub(in.start)
	for in.Anim != nil && in.cue < len(in.Anim.Cues) && in.Anim.Cues[in.cue].At <= t {
		in.Cues = append(in.Cues, in.Anim.Cues[in.cue])
		in.cue++
	}
	if skip || in.Anim == nil || t >= in.Anim.Len() {
		in.done = true
		if in.Anim != nil && len(in.Anim.Cues) > 0 { // VACE stops its song on the way out
			in.Cues = append(in.Cues, vace.Cue{At: t, Op: "stop"})
		}
		name := sq.Name
		h.Call(func(e *classicGame) { e.SequenceDone(name) })
		return in.last()
	}
	for i := range in.Anim.Frames {
		if t < in.Anim.Frames[i].Delay {
			return &in.Anim.Frames[i]
		}
		t -= in.Anim.Frames[i].Delay
	}
	return in.last()
}

func (in *SequenceUI) last() *vace.VaceFrame {
	if in.Anim == nil || len(in.Anim.Frames) == 0 {
		return nil
	}
	return &in.Anim.Frames[len(in.Anim.Frames)-1]
}
