package classic

import (
	"gobox/front"
	"gobox/game"
)

// Walls is v with the 3D view s.Walls holds: the wall faces, and each wall
// value's picture from art (Data.Art, read live) as View pixels or a DAX
// ref. v where s has none.
func Walls(v front.View, s game.State, art map[string]game.Picture) front.View {
	w := s.Walls
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
	v.Walls = f
	return v
}
