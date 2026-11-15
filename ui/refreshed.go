package ui

import (
	"image/color"
	"time"

	"gobox/game"
	"gobox/gpu"
)

// Camp's Fix is one moment of the front's: the screen fades to black at
// once (the player sees the command taken), the game does the whole fix
// unseen while it is black (game.Refreshed says it is done, with one
// sound), a box says what was restored and waits for a key, and the screen
// fades in again. A fix that ends otherwise (interrupted, nothing to fix)
// fades in on what the game shows.

// refreshedCue is the one sound of a Fix: the sound pack's recovery chime.
const refreshedCue = "recover"

// refreshFade is how long the screen takes to fade out, and in again.
const refreshFade = 400 * time.Millisecond

// refreshed is what a Fix just done restored, as its box says it.
type refreshed struct{ lines []string }

// newRefreshed is the box's lines for e.
func newRefreshed(e game.Refreshed) *refreshed {
	r := &refreshed{}
	if e.Healed {
		r.lines = append(r.lines, "HP restored.")
	}
	if e.Studied {
		r.lines = append(r.lines, "Spells memorized.")
	}
	return r
}

// fixPhase is where a Fix's moment is.
type fixPhase int

const (
	fixNone    fixPhase = iota
	fixFadeOut          // the screen darkens; the rest is sent once it is black
	fixWorking          // black: the game is fixing
	fixSaid             // black, the box up: any key or click goes on
	fixFadeIn           // the screen comes back
)

// fixMoment is the Fix under way.
type fixMoment struct {
	phase fixPhase
	at    time.Time // when the phase began
	rest  int       // the rest to send (Data.Rests' N)
	black bool      // fixWorking: a wholly black frame has been drawn, the game may work
	sent  bool      // fixWorking: the rest is with the game
	done  bool      // fixWorking: the game has answered it
	lines []string  // the box's
}

// startFix begins the moment for rest n (Data.Rests' Fix).
func (p *Presenter) startFix(n int) {
	if p.fix.phase == fixNone {
		p.fix = fixMoment{phase: fixFadeOut, at: now(), rest: n}
	}
}

// fixStep moves the moment on with time and with what the game answered.
func (p *Presenter) fixStep() {
	m := &p.fix
	switch m.phase {
	case fixFadeOut:
		if now().Sub(m.at) >= refreshFade {
			m.phase = fixWorking
		}
	case fixWorking:
		if !m.black { // this frame is the first wholly black one: the game works on the next, behind it
			m.black = true
			return
		}
		r := p.eng.r
		switch {
		case !m.sent: // the whole fix, at once (the game's Fix is done unseen)
			m.sent, r.refreshed = true, nil
			rest := m.rest
			p.call(func(e *classicGame) {
				e.do(game.CampRest{Kind: rest})
				p.fix.done = true
			})
		case !m.done:
		case r.refreshed != nil:
			m.phase, m.lines, r.refreshed = fixSaid, r.refreshed.lines, nil
		default: // no Refreshed (interrupted, nothing to fix, refused): back to what the game shows
			m.phase, m.at = fixFadeIn, now()
		}
	case fixFadeIn:
		if now().Sub(m.at) >= refreshFade {
			*m = fixMoment{}
		}
	}
}

// fixInput takes a key or click while a Fix's moment runs (none acts on
// the screen under it): on the box it goes on to the fade in. False: no
// moment runs.
func (p *Presenter) fixInput() bool {
	m := &p.fix
	if m.phase == fixNone {
		return false
	}
	if m.phase == fixSaid {
		m.phase, m.at = fixFadeIn, now()
	}
	return true
}

// dark is how black the screen is now (0 clear .. 1 black).
func (m *fixMoment) dark() float32 {
	t := min(float32(now().Sub(m.at))/float32(refreshFade), 1)
	switch m.phase {
	case fixFadeOut:
		return t
	case fixWorking, fixSaid:
		return 1
	case fixFadeIn:
		return 1 - t
	}
	return 0
}

// drawFix draws a Fix's moment over the screen drawn.
func (p *Presenter) drawFix(dst *gpu.Image, glyph func(byte) [8]byte) {
	if p.eng == nil || p.fix.phase == fixNone {
		return
	}
	p.fixStep()
	m := &p.fix
	b := dst.Bounds()
	gpu.DrawFilledRect(dst, 0, 0, float32(b.Dx()), float32(b.Dy()), color.NRGBA{A: uint8(255 * m.dark())})
	if m.phase == fixSaid && len(m.lines) > 0 {
		k := scale(dst)
		rects := p.menuRects // the box is no menu: the rows of the one under it stay the pointer's
		defer func() { p.menuRects = rects }()
		p.drawMenu(dst, glyph, "", m.lines, make([]byte, len(m.lines)), nil, -1, k, b.Dx()/2, (b.Dy()-int(k)*8*(len(m.lines)+4))/2)
	}
}
