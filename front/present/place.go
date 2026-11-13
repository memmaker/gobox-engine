package present

import (
	"image"
	"strconv"

	"gobox/blobber"
	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// Place is v with the view window and the party's place from where: the window
// (numbered as front.Window), the area, the square and the way it faces.
func Place(v front.View, where game.Where, window game.Window) front.View {
	a, _ := strconv.Atoi(where.Area)
	v.Window, v.Area = front.Window(window), front.AreaID(a)
	v.X, v.Y, v.Facing = where.X, where.Y, front.Dir(where.Facing/2&3)
	if q := where.Square; q != nil {
		v.X, v.Y = q.X, q.Y
	}
	return v
}

// where is the party's place in the state up (zero with no game).
func (p *Presenter) where() game.Where {
	if p.eng == nil {
		return game.Where{}
	}
	return play.CommonOf(p.state()).Where
}

// squareOf is the square the party stands on: Square where X, Y is another map's.
func squareOf(w game.Where) image.Point {
	if q := w.Square; q != nil {
		return *q
	}
	return image.Pt(w.X, w.Y)
}

// cameraOf is the 3D view's camera at w.
func cameraOf(w game.Where) blobber.Camera {
	q := squareOf(w)
	return blobber.Camera{X: q.X, Y: q.Y, Dir: w.Facing / 2 & 3}
}

// window is what the view window shows (the first, the 3D view, with no game).
func (p *Presenter) window() game.Window {
	if p.eng == nil {
		return game.WindowScene
	}
	return windowOf(p.state())
}

// open is the party's square's sides (N E S W) a step may go through (none with no map).
func (p *Presenter) open() (o [4]bool) {
	if p.eng == nil {
		return o
	}
	if m := p.eng.r.mapOf(p.state()); m != nil {
		return openSides(m, p.where())
	}
	return o
}

// flags are the plain status flags of the state up (none with no game).
func (p *Presenter) flags() *game.Flags {
	if p.eng == nil {
		return &game.Flags{}
	}
	s := p.state()
	return p.eng.r.flagsOf(s, partyOf(s))
}

// minimap is the automap of the state up, and whether it is a 16x16 map to draw.
func (p *Presenter) minimap() (game.Automap, bool) {
	if p.eng == nil {
		return game.Automap{}, false
	}
	a := p.eng.r.automapOf(p.state())
	return a, a.Size == image.Pt(16, 16)
}
