package present

import (
	"gobox/blobber"
	"gobox/front"
)

// Free reports whether No Grid walking has the move keys: the party
// explores the 3D view (v) with no menu up.
func (p *Presenter) Free(v front.View) bool {
	return blobber.Settings.NoGrid && p.active && p.view != nil && p.shown != "" && !p.menu && !p.view.OptionsOpen() &&
		v.Exploring && v.Window != front.WinAreaMap
}

// ViewClick is a click on No Grid's view, taken on its release (no drag):
// as a click on the view without No Grid.
func (p *Presenter) ViewClick(x, y int) {
	if v := p.shownView; p.Free(v) {
		p.exploreClick(x, y, v)
	}
}

// FreeWalk moves No Grid's camera by one tick of in, and sends the step or
// turn it asks for: a step aside or back keeps the facing, a turn is quiet
// (the mouse turns a lot).
func (p *Presenter) FreeWalk(in blobber.FreeInput, v front.View) {
	if p.view == nil {
		return
	}
	if !blobber.Settings.NoGrid || p.shown == "" {
		p.view.FreeOff()
		return
	}
	ready := p.Free(v) && len(p.inv.acts) == 0
	at := blobber.Camera{X: v.X, Y: v.Y, Dir: int(v.Facing)}
	c := p.view.Walk(at, v.Open, in, ready)
	if !ready {
		return
	}
	steps := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	for d, s := range steps {
		if s == [2]int{c.X - at.X, c.Y - at.Y} {
			p.send([4]front.ActKind{front.ActForward, front.ActStrafeRight, front.ActBackward, front.ActStrafeLeft}[(d-at.Dir+4)%4], 0)
			return
		}
	}
	switch (c.Dir - at.Dir + 4) % 4 {
	case 1, 2:
		p.send(front.ActTurnRight, 1)
	case 3:
		p.send(front.ActTurnLeft, 1)
	}
}
