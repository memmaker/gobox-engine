package engine

import (
	"gobox/game"
	"gobox/rules"
)

// mapWalls is the 3D view's walls: every square's wall values by side
// (WallAt) and each value's remastered texture from the game's
// GameTables.Walls. None off the 3D view or for a game without the table
// (POR: its remastered area maps draw the view).
func (e *Engine) mapWalls() game.MapWalls {
	wt := e.game().Walls
	if wt == nil || e.viewMode != 1 {
		return game.MapWalls{}
	}
	pool := map[int]string{}
	block, chunk := -1, 0
	for s := range rules.WallSlots { // a slot without its own block holds the next chunk of the block before it
		if b := e.mapPairA[s]; b >= 0 && b != walldefDefault {
			block, chunk = b, 1
		} else if block >= 0 {
			chunk++
		}
		for _, t := range wt.Walls {
			if t.Set == e.set && walldefRecorded(t.Block) == block && t.Chunk == chunk {
				pool[s*rules.SetWalls+t.Wall] = t.Texture
			}
		}
	}
	if len(pool) == 0 {
		return game.MapWalls{}
	}
	w := game.MapWalls{Side: make([][][4]int, rules.AreaSize), Pool: pool, Floor: wt.Floor, Sky: wt.Sky}
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
