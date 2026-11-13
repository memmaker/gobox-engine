package present

import (
	"gobox/front"
	"gobox/game"
)

// Walls is v with the 3D view w: the wall faces, and each wall
// value's picture from art (Data.Art, read live) as View pixels or a DAX
// ref. v where w is nil.
func Walls(v front.View, w *game.Walls, art map[string]game.Picture) front.View {
	if w == nil {
		return v
	}
	f := &front.Walls{Key: w.Key, Side: w.Side, Pool: w.Pool, Floor: w.Floor, Sky: w.Sky,
		FloorRGBA: w.FloorRGBA, CeilRGBA: w.CeilRGBA, Sun: w.Sun, SunAz: w.SunAz, SunEl: w.SunEl, Mountains: w.Mountains}
	if w.Fronts != nil {
		f.FrontPix = make(map[int]front.Picture, len(w.Fronts))
		for k, name := range w.Fronts {
			p := art[name]
			fp := pix(p)
			if fp.Pix == nil && p.File != "" {
				fp = front.Picture{File: p.File, Block: p.Block, Frame: p.Frame}
			}
			f.FrontPix[k] = fp
		}
	}
	if b := w.Backdrop; b != nil {
		fb := pix(*b)
		if fb.Pix == nil {
			fb = front.Picture{File: b.File, Block: b.Block, Frame: b.Frame}
		}
		f.Backdrop = &fb
	}
	v.Walls = f
	return v
}
