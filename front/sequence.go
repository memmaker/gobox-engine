// sequence.go plays a full-screen animation (View.Sequence: VACE's
// .ACF/.BMA, gobox/dax/vace) and ends it (SequenceDone) when it runs out
// or on a key, click or pad button.

package front

import (
	"time"

	"gobox/dax/vace"
)

// Sequence is the animation State.Sequence names: Anim its files' base name in Dir.
type Sequence struct {
	Name, Anim, Dir string
}

// SequenceUI is the animation being played.
type SequenceUI struct {
	Name  string
	Anim  *vace.Anim // nil: its files did not decode (it ends at once)
	start time.Time
	done  bool // SequenceDone sent
}

// SequenceStep plays sq (nil: none) this frame on in and returns the
// frame to show (nil: none); skip: a key, click or button came.
func SequenceStep(h Host, in *SequenceUI, sq *Sequence, skip bool) *vace.Frame {
	if sq == nil {
		*in = SequenceUI{}
		return nil
	}
	now := time.Now()
	if in.Name != sq.Name || in.start.IsZero() {
		a, _ := vace.Load(sq.Dir, sq.Anim)
		*in = SequenceUI{Name: sq.Name, Anim: a, start: now}
	}
	if in.done {
		return in.last()
	}
	t := now.Sub(in.start)
	if skip || in.Anim == nil || t >= in.Anim.Len() {
		in.done = true
		name := sq.Name
		h.Call(func(e Game) { e.SequenceDone(name) })
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

func (in *SequenceUI) last() *vace.Frame {
	if in.Anim == nil || len(in.Anim.Frames) == 0 {
		return nil
	}
	return &in.Anim.Frames[len(in.Anim.Frames)-1]
}
