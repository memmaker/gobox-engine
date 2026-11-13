package engine

import (
	"gobox/game"
	"gobox/rules"
)

// wallChunk is the WALLDEF chunk a wall slot holds: chunk Chunk (1-based) of
// block Block (as recorded); Block -1: none.
type wallChunk struct{ Block, Chunk int }

// wallChunks is the chunk each wall slot holds: a slot without its own
// block holds the next chunk of the block before it.
func (e *Engine) wallChunks() [rules.WallSlots]wallChunk {
	var out [rules.WallSlots]wallChunk
	block, chunk := -1, 0
	for s := range rules.WallSlots {
		if b := e.mapPairA[s]; b >= 0 && b != walldefDefault {
			block, chunk = b, 1
		} else if block >= 0 {
			chunk++
		}
		out[s] = wallChunk{block, chunk}
	}
	return out
}

// mapWalls is the 3D view's walls: every square's wall values by side
// (WallAt), each value's remastered texture from the game's
// GameTables.Walls and the area's scenery (WallTextures.Areas, by set and
// GEO block). None off the 3D view or for a game without the table.
func (e *Engine) mapWalls() game.MapWalls {
	wt := e.game().Walls
	if wt == nil || e.viewMode != 1 {
		return game.MapWalls{}
	}
	pool := map[int]string{}
	for s, c := range e.wallChunks() {
		for _, t := range wt.Walls {
			if t.Set == e.set && walldefRecorded(t.Block) == c.Block && t.Chunk == c.Chunk {
				pool[s*rules.SetWalls+t.Wall] = t.Texture
			}
		}
	}
	if len(pool) == 0 {
		return game.MapWalls{}
	}
	w := game.MapWalls{Side: make([][][4]int, rules.AreaSize), Pool: pool, Floor: wt.Floor, Sky: wt.Sky}
	if a := wt.Area(e.set, e.worldWord(rules.WordGeo)); a != nil {
		if a.Floor != "" {
			w.Floor = a.Floor
		}
		if a.Sky != "" {
			w.Sky = a.Sky
		}
		w.Scenery = a.Scenery
	}
	for y := range w.Side {
		w.Side[y] = make([][4]int, rules.AreaSize)
		for x := range w.Side[y] {
			for side := range w.Side[y][x] {
				w.Side[y][x][side] = int(e.WallAt(x, y, 2*side))
			}
		}
	}
	return w
}
