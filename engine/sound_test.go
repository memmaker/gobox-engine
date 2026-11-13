package engine

import (
	"fmt"
	"testing"
)

// fakeSound records the driver calls (Sound).
type fakeSound struct {
	calls []string
	card  int
}

func (f *fakeSound) log(format string, a ...any) {
	f.calls = append(f.calls, fmt.Sprintf(format, a...))
}
func (f *fakeSound) Effect(n int, quiet bool) { f.log("Effect %d %v", n, quiet) }
func (f *fakeSound) CutAfter(frames int)      { f.log("CutAfter %d", frames) }
func (f *fakeSound) Wait(n int)               { f.log("Wait %d", n) }
func (f *fakeSound) Reset()                   { f.log("Reset") }
func (f *fakeSound) SetCard(c int)            { f.card = c; f.log("SetCard %d", c) }
func (f *fakeSound) Song(n, next int, again bool) bool {
	f.log("Song %d %d %v", n, next, again)
	return false
}
func (f *fakeSound) AmigaSong(n int) { f.log("AmigaSong %d", n) }
func (f *fakeSound) SetSoundtrack(on bool, n, music int) {
	f.log("SetSoundtrack %v %d %d", on, n, music)
}
func (f *fakeSound) Track() (bool, int) { return false, 0 }
func (f *fakeSound) Sounding() bool     { return false }

func soundEngine(t *testing.T, card byte) (*Engine, *fakeSound) {
	e := &Engine{Cfg: Config{Sound: card}}
	e.initVM()
	f := &fakeSound{}
	e.SetSound(f)
	f.calls = nil
	return e, f
}

// In the cut mode a sound is cut a tenth of a second on; after soundWait(0)
// it plays on.
func TestSoundCut(t *testing.T) {
	e, f := soundEngine(t, 'T')
	e.PlaySound(sndSpell)
	e.soundWait(0)
	e.PlaySound(sndHit)
	want := fmt.Sprint([]string{"Effect 2 false", fmt.Sprintf("CutAfter %d", sndRate/10), "Wait 0", "Effect 3 false"})
	if got := fmt.Sprint(f.calls); got != want {
		t.Errorf("calls %s, want %s", got, want)
	}
}

// With "Step sounds" off a step or a turn is silent.
func TestNoSteps(t *testing.T) {
	e, f := soundEngine(t, 'P')
	e.NoSteps.Store(true)
	e.PlaySound(sndStep)
	if len(f.calls) != 0 {
		t.Fatalf("a step sounds with Step sounds off: %v", f.calls)
	}
}

// soundWait = 08AE:01A5 (n): 0 hooks the timer interrupt so sounds play
// while the caller animates; n > 0 unhooks it (when it was hooked) and sets
// the cut-after-0.1 s mode used by PlaySound.
func (e *Engine) soundWait(n int) {
	s := &e.snd
	if s.d == nil || s.typ == sndNone {
		return
	}
	s.d.Wait(n)
	s.wait = n
}
