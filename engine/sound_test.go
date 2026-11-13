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

// Ctrl-S mutes the card and stops the song, again restores the card.
func TestSoundToggle(t *testing.T) {
	e, f := soundEngine(t, 'T')
	e.Cfg.SaveDir = t.TempDir()
	e.soundToggle()
	if e.snd.typ != sndNone || f.card != sndNone {
		t.Fatalf("Ctrl-S: typ %d card %d", e.snd.typ, f.card)
	}
	e.soundToggle()
	if e.snd.typ != sndTandy || f.card != sndTandy {
		t.Fatalf("Ctrl-S again: typ %d card %d", e.snd.typ, f.card)
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
