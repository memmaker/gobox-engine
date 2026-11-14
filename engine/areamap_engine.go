package engine

import "gobox/rules"

// DrawAreaMap = ovr28:0000 (x, y): the area map in the view window. The
// presenter draws it from the State's Automap.Terrain.
func (e *Engine) DrawAreaMap() { e.setWin(WinAreaMap) }

// terrainPix is the picture of the terrain tile v (bit 7: the second set);
// empty for a tile the sets lack.
func (e *Engine) terrainPix(v byte) Picture {
	im, f := e.sqrpaci[0], int(v)
	if v > 0x7F {
		im, f = e.sqrpaci[1], int(v&0x7F)
	}
	if im == nil || f >= im.Count {
		return Picture{}
	}
	return e.picPix(im, f, false, nil)
}

// areaMapInit = ovr28:0207: the three frames of the area map screen.
func (e *Engine) areaMapInit() {
	e.newPage()
	e.blank(1, 1, 15, 15)
	e.blank(17, 1, winRight, 15)
	e.blank(1, msgTop, winRight, winBottom)
}

// areaMapIcons = 00BA:2B5E: switch to the area map view and draw it.
func (e *Engine) areaMapIcons() {
	e.setWorldWord(rules.WordBeginFlag, 3)
	e.areaMapInit()
	e.Redraw()
}
