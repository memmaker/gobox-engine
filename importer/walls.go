package importer

import (
	"embed"
	"fmt"
	"strings"

	"gobox/adventure"
	"gobox/rules"
)

// walls/<id>.txt (from the per-game remasters' wall indexes) names each
// WALLDEF wall's remastered texture, a line a wall: SET BLOCK CHUNK WALL
// TEXTURE (CHUNK 1-based, WALL 1..rules.SetWalls).
//
//go:embed walls/*.txt
var wallFiles embed.FS

// wallTextures is game id's wall textures under the flats floor and sky.
// The tables are embedded: a malformed line is a build defect (TestWallTextures).
func wallTextures(id, floor, sky string) *adventure.WallTextures {
	b, err := wallFiles.ReadFile("walls/" + id + ".txt")
	if err != nil {
		panic(err)
	}
	w := &adventure.WallTextures{Floor: floor, Sky: sky}
	for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		var t adventure.WallTexture
		if _, err := fmt.Sscan(l, &t.Set, &t.Block, &t.Chunk, &t.Wall, &t.Texture); err != nil || t.Wall < 1 || t.Wall > rules.SetWalls || t.Chunk < 1 {
			panic(fmt.Sprintf("walls/%s.txt: %q: %v", id, l, err))
		}
		w.Walls = append(w.Walls, t)
	}
	return w
}

// The games' 3D view flats (the pool series' floor_<x>, ceiling_<x>):
// cotab's cobblestones (SKY.DAX 252) under its stone roof, pod's rock cave
// (SKYGRND 4); sotsb none, the plain colours.
// ponytail: one sky a game; the originals' outdoor skies (cotab's by sky
// colour, pod's and sotsb's open-air areas) need the engine's sky fact.
const (
	cotabFloor, cotabSky = "cobble", "cotab_stone"
	podFloor, podSky     = "pod_rock", "pod_rock"
)
