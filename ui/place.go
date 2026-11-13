package ui

import (
	"image"

	"gobox/game"
	"gobox/game/play"
)

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
func cameraOf(w game.Where) Camera {
	q := squareOf(w)
	return Camera{X: q.X, Y: q.Y, Dir: w.Facing / 2 & 3}
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
	r := p.eng.r
	return p.lay.flags.get(flagsKey{p.eng.G.Version(), r.title}, func() *game.Flags { s := p.state(); return r.flagsOf(s, partyOf(s)) })
}

// minimap is the automap of the state up and its squares, when it is a 16x16 map to draw.
func (p *Presenter) minimap() (game.Automap, *MapSquares) {
	if p.eng == nil {
		return game.Automap{}, nil
	}
	d := p.eng.r.derived()
	return d.automap, d.minimap
}

// walls is the 3D view the state shows (nil: none).
func (p *Presenter) walls() *game.Walls {
	if p.eng == nil {
		return nil
	}
	return p.eng.r.derived().walls
}
