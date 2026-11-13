// title.go plays the title (Intro, ovr02:0626 title_screen): TITLE
// picture 1 five seconds, with the full intro picture 2 five seconds with
// its lightning, then the credits ten; a key or click cuts each short (and
// a lightning flash one step). The game draws (TitlePicture,
// TitleCredits; the presenter flashes from IntroUI.Flash) and ends it (TitleDone).

package present

import (
	"gobox/game"
	"math/rand/v2"
	"slices"
	"time"
)

// The title's phases.
const (
	IntroNone = iota // not started
	IntroFirst
	IntroSecond
	IntroCredits = 1 << 16 // after the last picture (the pictures are Phase 1, 2, ...)
)

// IntroUI is the title being played.
type IntroUI struct {
	Phase     int
	end       time.Time // the phase's end
	Flash     []int     // the lightning's colours left, a step each
	ms        []int     // ... and each step's length
	step      time.Time // the flash step's end
	cut       bool      // a key cut a phase short since the last frame
	acc       int       // the lightning's chance, as the original's counter
	NoCredits bool      // the presenter's: the game has no credits page (its credits are a picture step), so no credits step
}

// IntroStep plays the title (nil: none waits) this frame on in; key: a
// key or click came.
func IntroStep(h RuleHost, in *IntroUI, intro *game.Intro, key bool) {
	if intro == nil {
		*in = IntroUI{}
		return
	}
	now := time.Now()
	if !key {
		in.cut = false
	} else if in.cut { // one key cuts one phase, however often it is delivered
		return
	}
	pics := intro.Pictures
	if pics == nil {
		pics = []int{1}
		if intro.Full {
			pics = []int{1, 2}
		}
	}
	if len(in.Flash) > 0 { // Delay: the step goes on until its time is up or a key
		if !key && now.Before(in.step) {
			return
		}
		in.Flash, in.ms = in.Flash[1:], in.ms[1:]
		if len(in.Flash) > 0 {
			in.step = now.Add(time.Duration(in.ms[0]) * time.Millisecond)
		}
		return
	}
	if in.Phase == IntroSecond && !key && intro.Lightning {
		if in.acc += rand.IntN(4); in.acc > 50000 {
			in.acc = 0
			for range rand.IntN(3) + 1 {
				in.Flash = append(in.Flash, 9, 11, 15, 11, 9)
				in.ms = append(in.ms, 150, 100, 50, 100, 150)
			}
			in.step = now.Add(time.Duration(in.ms[0]) * time.Millisecond)
			return
		}
	}
	if in.Phase != IntroNone && !key && now.Before(in.end) {
		return
	}
	in.cut = key
	switch {
	case in.Phase >= IntroFirst && in.Phase < len(pics) || in.Phase == IntroNone:
		in.Phase++
		n := pics[in.Phase-1]
		if n < 0 { // a credits page step
			in.end = now.Add(10 * time.Second)
			h.Call(func(e *classicGame) { e.TitleCredits(-n - 1) })
			break
		}
		in.end = now.Add(5 * time.Second)
		h.Call(func(e *classicGame) { e.TitlePicture(n) })
	case !in.NoCredits && in.Phase != IntroCredits && len(pics) > 1 && !slices.ContainsFunc(pics, func(n int) bool { return n < 0 }): // a lone picture ends the title
		in.Phase, in.end = IntroCredits, now.Add(10*time.Second)
		h.Call(func(e *classicGame) { e.TitleCredits(0) })
	default:
		h.Call(func(e *classicGame) { e.TitleDone() })
	}
}
