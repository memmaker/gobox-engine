// sequence.go plays a full-screen animation (Sequence: VACE's
// .ACF/.BMA, gobox/dax/vace) and ends it (SequenceDone) when it runs out
// or on a key, click or pad button.

package ui

import (
	"gobox/adventure"
	"time"
)

// SequenceUI is the animation being played.
type SequenceUI struct {
	Name  string
	Anim  *Anim // nil: its files did not decode (it ends at once)
	start time.Time
	done  bool // SequenceDone sent
	cue   int  // Anim.Cues fired so far
	// Cues are the music cues (Anim.Cues) that came due in the last step; a
	// "stop" ends it when the animation ends. The presenter plays them.
	Cues []Cue
}

// seqAnim is the animation named name (game.Scene.Sequence, a menu item's):
// the leading StepAnim of the adventure's script it names (nil: none).
func seqAnim(name string) *adventure.SequenceStep {
	switch name {
	case adventure.SeqIntro:
		return startSeq.LeadingAnim()
	case adventure.SeqEnding:
		return endSeq.LeadingAnim()
	}
	return nil
}

// SequenceStep plays the animation name ("" or none such: none) this frame
// on in and returns the frame to show (nil: none); skip: a key, click or
// button came.
func SequenceStep(h RuleHost, in *SequenceUI, name string, skip bool) *VaceFrame {
	sq := seqAnim(name)
	if sq == nil {
		*in = SequenceUI{}
		return nil
	}
	now := time.Now()
	if in.Name != name || in.start.IsZero() {
		a, _ := Load(filesDir, sq.Anim)
		*in = SequenceUI{Name: name, Anim: a, start: now}
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
			in.Cues = append(in.Cues, Cue{At: t, Op: "stop"})
		}
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

func (in *SequenceUI) last() *VaceFrame {
	if in.Anim == nil || len(in.Anim.Frames) == 0 {
		return nil
	}
	return &in.Anim.Frames[len(in.Anim.Frames)-1]
}
