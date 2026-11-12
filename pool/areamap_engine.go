package pool

import "gobox/core"

// DrawAreaMap = ovr28:0000 (x, y): the 5x5 tiles of the area map starting at
// map square (x,y), each drawn as a 3x3 cell tile. Landmark squares listed in
// the DS:0C30 table (6 bytes each: x, y, icon) replace the tile, some only
// while the area record flags say they are revealed.
func (e *Engine) DrawAreaMap(x, y int) {
	e.snap.win = WinAreaMap
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			e.drawTile(3*col, 3*row, byte(e.ruleset().Check(core.AreaTileIs{X: int32(col + x), Y: int32(row + y)}, nil).Value))
		}
	}
}

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
	return e.picPix(im, f, false, -1)
}

// drawAreaMapParty = ovr27:0000: the party marker on the area map. The
// marker square is the party position relative to the shown window, pinned
// to the border squares; the icon animates through 6 poses (DS:6DDC).
func (e *Engine) drawAreaMapParty() {
	v := int8(int(ViewModeShift[e.viewMode]) + e.worldWord(core.WordMapX))
	switch {
	case v < 2:
	case v < 0x2A:
		v = 2
	case v < 0x2B:
		v = 3
	default:
		v = 4
	}
	var y byte
	switch w := e.worldWord(core.WordMapY); {
	case w < 2:
		y = byte(w)
	case w < 0x22:
		y = 2
	case w < 0x23:
		y = 3
	default:
		y = 4
	}
	c := e.areaIconTick
	e.drawIcon(int(v), int(y), byte(e.facing), c%2, c/2+9)
	e.areaIconTick = (c + 1) % 6
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
	e.setWorldWord(core.WordBeginFlag, 3)
	e.areaMapInit()
	e.Redraw()
	e.drawAreaMapParty()
}
