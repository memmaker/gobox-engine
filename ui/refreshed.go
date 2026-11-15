package ui

import (
	"image/color"
	"time"

	"gobox/game"
	"gobox/gpu"
)

// Camp's Fix is done by the game at once and unseen (game.Refreshed). The
// front shows it as one moment: the screen fades to black, a box says what
// was restored, the screen fades in again; one sound plays (refreshedCue).

// refreshedCue is the one sound of a Fix: the sound pack's recovery chime.
const refreshedCue = "recover"

// The moment's times: the fade out, the words held on black, the fade in.
const refreshFade, refreshHold = 400 * time.Millisecond, 1600 * time.Millisecond

// refreshed is a Fix just done: what it restored, and when.
type refreshed struct {
	lines []string
	at    time.Time
}

// newRefreshed is the moment for e, begun at at.
func newRefreshed(e game.Refreshed, at time.Time) *refreshed {
	r := &refreshed{at: at}
	if e.Healed {
		r.lines = append(r.lines, "HP restored.")
	}
	if e.Studied {
		r.lines = append(r.lines, "Spells memorized.")
	}
	return r
}

// dark is how black the screen is at t (0 clear .. 1 black), and whether
// the words show; over: the moment has passed.
func (r *refreshed) dark(t time.Time) (a float32, words, over bool) {
	switch d := t.Sub(r.at); {
	case d < refreshFade:
		return float32(d) / float32(refreshFade), false, false
	case d < refreshFade+refreshHold:
		return 1, true, false
	case d < 2*refreshFade+refreshHold:
		return 1 - float32(d-refreshFade-refreshHold)/float32(refreshFade), false, false
	}
	return 0, false, true
}

// drawRefreshed draws a Fix's moment over the screen drawn, and forgets it
// once it has passed.
func (p *Presenter) drawRefreshed(dst *gpu.Image, glyph func(byte) [8]byte) {
	if p.eng == nil || p.eng.r.refreshed == nil {
		return
	}
	r := p.eng.r.refreshed
	a, words, over := r.dark(now())
	if over {
		p.eng.r.refreshed = nil
		return
	}
	b := dst.Bounds()
	gpu.DrawFilledRect(dst, 0, 0, float32(b.Dx()), float32(b.Dy()), color.NRGBA{A: uint8(255 * a)})
	if words && len(r.lines) > 0 {
		k := scale(dst)
		rects := p.menuRects // the box is no menu: the rows of the one under it stay the pointer's
		defer func() { p.menuRects = rects }()
		p.drawMenu(dst, glyph, "", r.lines, make([]byte, len(r.lines)), nil, -1, k, b.Dx()/2, (b.Dy()-int(k)*8*(len(r.lines)+4))/2)
	}
}
