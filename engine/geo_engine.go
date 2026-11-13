package engine

import (
	"gobox/asset"
	"gobox/rules"
)

// gmap is the area map; por's edges are always solid (see offMapSolid).
func (e *Engine) gmap() *rules.Map { e.gs.geo.SolidEdges = true; return &e.gs.geo }

// WallAt = ovr30:047B (x, y, dir): wall value 0..15 on that side of the cell.
func (e *Engine) WallAt(x, y, dir int) byte { return e.gmap().WallAt(x, y, dir) }

// DoorAt = ovr30:0350 (x, y, dir): the 2-bit door state on that side.
// Off the map it is 0; on the map it defaults to 1 for an open side (no wall).
func (e *Engine) DoorAt(x, y, dir int) byte { return e.gmap().DoorAt(x, y, dir) }

// MapCell = ovr30:057A (x, y): the cell byte (plane 2).
func (e *Engine) MapCell(x, y int) byte { return e.gmap().MapCell(x, y) }

// Draw3DView = ovr30:05FD (x, y, dir): the first-person view from cell
// (x,y). The remaster draws it (the page's view window stays black under
// it); here the window is set to it.
func (e *Engine) Draw3DView(x, y, dir int) {
	e.snap.win = WinThreeD
	e.erase(InnerWindow.Pixels())
}

// LoadWalldef = ovr30:0E08 (block, slot): WALLDEF<n>.DAX block into wall
// slot 1..3. Only the slot's block is recorded (the save file keeps it);
// the walls are the remaster's.
func (e *Engine) LoadWalldef(block, slot int) {
	if !rules.ValidWallSlot(slot) {
		return
	}
	if block == 0xA { // ovr30:1061: block 10 is recorded as 0 (quirk)
		block = 0
	}
	e.mapPairA[slot-1], e.mapPairB[slot-1] = block, slot
}

// LoadGeo = ovr30:10C7 (n): GEO<n>.DAX block n into the map planes.
func (e *Engine) LoadGeo(n int) {
	b := e.gdata().geoBlock(e.set.File("GEO"), n)
	if len(b) != 0 && len(b) != geoSize {
		e.PromptRow24(asset.EGAWhite, "Unable to load geo in Load3DMap.")
	}
	if len(b) >= geoSize { // a longer block is cut to its size
		if m, err := rules.ParseMap(b[:geoSize]); err == nil {
			e.gs.geo.Plane = m.Plane
		}
	}
	e.setWorldWord(rules.WordGeo, n)
	e.reopenDoors()
}

const geoSize = rules.Size // GEO block: 2-byte header + 4 planes

type geoState struct {
	geo      rules.Map // DS:6A5C
	areaKind byte      // DS:84DC: 0 or 0xA = map edges are solid
}
