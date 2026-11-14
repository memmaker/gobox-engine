package engine

import (
	"gobox/game"
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
	e.setWin(WinThreeD)
	e.unwrite(InnerWindow.Pixels())
}

// LoadWalldef = ovr30:0E08 (block, slot): WALLDEF<n>.DAX block into wall
// slot 1..3. A block of several wall sets (chunks) fills the slots from
// slot on, each cleared, and only slot records the block (por-re view3d.pas
// LoadWallSet: wallSetCache[..].wallDefId := -1, then [wsSlot] := fileId;
// the save file keeps them); the walls are the remaster's.
func (e *Engine) LoadWalldef(block, slot int) {
	if !rules.ValidWallSlot(slot) {
		return
	}
	chunks := 1
	if w, _ := rules.ParseWalldef(e.gdata().geoBlock(e.set.File("WALLDEF"), block)); len(w.Sets) > 1 {
		chunks = len(w.Sets)
	}
	for s := slot; s < slot+chunks && rules.ValidWallSlot(s); s++ {
		e.mapPairA[s-1], e.mapPairB[s-1] = -1, -1
	}
	e.mapPairA[slot-1], e.mapPairB[slot-1] = walldefRecorded(block), slot
}

// walldefRecorded is the block LoadWalldef records for block: ovr30:1061
// records block 10 as 0 (quirk).
func walldefRecorded(block int) int {
	if block == 0xA {
		return 0
	}
	return block
}

// LoadGeo = ovr30:10C7 (n): GEO<n>.DAX block n into the map planes.
func (e *Engine) LoadGeo(n int) {
	b := e.gdata().geoBlock(e.set.File("GEO"), n)
	if len(b) != 0 && len(b) != geoSize {
		e.PromptRow24(game.InkWhite, "Unable to load geo in Load3DMap.")
	}
	if len(b) >= geoSize { // a longer block is cut to its size
		if m, err := rules.ParseMap(b[:geoSize]); err == nil {
			e.gs.geo.Plane = m.Plane
		}
	}
	e.setWorldWord(rules.WordGeo, n)
	e.reopenDoors()
}

const geoSize = rules.GeoSize // GEO block: 2-byte header + 4 planes

type geoState struct {
	geo      rules.Map // DS:6A5C
	areaKind byte      // DS:84DC: 0 or 0xA = map edges are solid
}
