package ui

import (
	"gobox/game"
)

// Free reports whether No Grid walking has the move keys: the party
// explores the 3D view with no menu up.
func (p *Presenter) Free() bool {
	return Settings.NoGrid && p.active && p.view != nil && p.shown != "" && !p.menu && !p.optionsOpen() &&
		p.exploring() && p.window() != game.WindowMap
}

// ViewClick is a click on No Grid's view, taken on its release (no drag):
// as a click on the view without No Grid.
func (p *Presenter) ViewClick(x, y int) {
	if p.Free() {
		p.exploreClick(x, y)
	}
}

// FreeWalk moves No Grid's camera by one tick of in, and sends the step or
// turn it asks for: a step aside or back keeps the facing, a turn is quiet
// (the mouse turns a lot).
func (p *Presenter) FreeWalk(in FreeInput) {
	if p.view == nil {
		return
	}
	if !Settings.NoGrid || p.shown == "" {
		p.view.FreeOff()
		return
	}
	ready := p.Free() && len(p.inv.acts) == 0
	at := cameraOf(p.where())
	c := p.view.Walk(at, p.open(), in, ready)
	if !ready {
		return
	}
	steps := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	for d, s := range steps {
		if s == [2]int{c.X - at.X, c.Y - at.Y} {
			p.send([4]ActKind{ActForward, ActStrafeRight, ActBackward, ActStrafeLeft}[(d-at.Dir+4)%4], 0)
			return
		}
	}
	switch (c.Dir - at.Dir + 4) % 4 {
	case 1, 2:
		p.send(ActTurnRight, 1)
	case 3:
		p.send(ActTurnLeft, 1)
	}
}
