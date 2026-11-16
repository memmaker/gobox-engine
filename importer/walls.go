package importer

import (
	"cmp"
	"embed"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"gobox/adventure"
	"gobox/game"
	"gobox/rules"
)

// walls/<id>.txt (from the per-game remasters' wall indexes) names each
// WALLDEF wall's remastered texture, a line a wall: SET BLOCK CHUNK WALL
// TEXTURE (CHUNK 1-based, WALL 1..rules.SetWalls).
//
//go:embed walls/*.txt
var wallFiles embed.FS

// solidWalls are the pool textures that look like a solid wall
// (walls/solid.txt, a name a line; Felix's marking), for every game.
func solidWalls() map[string]bool {
	b, err := wallFiles.ReadFile("walls/solid.txt")
	if err != nil {
		panic(err)
	}
	solid := map[string]bool{}
	for _, n := range strings.Fields(string(b)) {
		solid[n] = true
	}
	return solid
}

// wallTextures is game id's wall textures under the flats floor and sky.
// The tables are embedded: a malformed line is a build defect (TestWallTextures).
func wallTextures(id, floor, sky string) *adventure.WallTextures {
	b, err := wallFiles.ReadFile("walls/" + id + ".txt")
	if err != nil {
		panic(err)
	}
	w := &adventure.WallTextures{Floor: floor, Sky: sky}
	solid := solidWalls()
	for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		var t adventure.WallTexture
		if _, err := fmt.Sscan(l, &t.Set, &t.Block, &t.Chunk, &t.Wall, &t.Texture); err != nil || t.Wall < 1 || t.Wall > rules.SetWalls || t.Chunk < 1 {
			panic(fmt.Sprintf("walls/%s.txt: %q: %v", id, l, err))
		}
		t.Solid = solid[t.Texture]
		w.Walls = append(w.Walls, t)
	}
	if b, err := wallFiles.ReadFile("walls/" + id + "_areas.txt"); err == nil {
		w.Areas = areaLooks(id, string(b))
	}
	if b, err := wallFiles.ReadFile("walls/" + id + "_backs.txt"); err == nil {
		w.Backs = backLooks(id, string(b))
	}
	return w
}

// backNoneWord is the block field of a line that gives the set's BackNone look (br2_backs.txt: buckrogers2-go rules/gamewalls.go
// wallsState's default for a view with no BACK picture, "br1_city" and "sky_br1_city"; BackFlats are its other lines).
const backNoneWord = "none"

// backLooks reads walls/<id>_backs.txt, the flats of each BACK picture, a
// line a picture: SET BLOCK FLOOR SKY (buckrogers-go rules/gamewalls.go
// backFlats, blocks with the same picture share names). Each set also gets
// its BackNone look, the one of a view without a picture (the original's sky
// fill over the ground): its first sky_* pair, else its first pair
// (gamewalls.go diskSky).
func backLooks(id, text string) []adventure.BackLook {
	var out []adventure.BackLook
	none := map[rules.Set]adventure.BackLook{}
	explicit := map[rules.Set]bool{}
	for _, l := range strings.Split(strings.TrimSpace(text), "\n") {
		b := adventure.BackLook{}
		var block string
		if _, err := fmt.Sscan(l, &b.Set, &block, &b.Floor, &b.Sky); err != nil {
			panic(fmt.Sprintf("walls/%s_backs.txt: %q: %v", id, l, err))
		}
		if block == backNoneWord { // the set's look without a picture, where the game's own default is not the first sky pair
			b.Block = adventure.BackNone
			none[b.Set], explicit[b.Set] = b, true
			continue
		}
		if n, err := strconv.Atoi(block); err != nil || n < 0 {
			panic(fmt.Sprintf("walls/%s_backs.txt: %q: block", id, l))
		} else {
			b.Block = n
		}
		out = append(out, b)
		if n, ok := none[b.Set]; !explicit[b.Set] && (!ok || (!strings.HasPrefix(n.Sky, "sky_") && strings.HasPrefix(b.Sky, "sky_"))) {
			none[b.Set] = adventure.BackLook{Set: b.Set, Block: adventure.BackNone, Floor: b.Floor, Sky: b.Sky}
		}
	}
	for _, n := range none {
		out = append(out, n)
	}
	slices.SortFunc(out, func(a, b adventure.BackLook) int {
		return cmp.Or(cmp.Compare(a.Set, b.Set), cmp.Compare(a.Block, b.Block))
	})
	return out
}

// areaLooks reads walls/<id>_areas.txt, the areas' own looks, a line a
// thing: SET GEO then one of
//
//	floor NAME / ceiling NAME                  the area's flats (pool floor_NAME, ceiling_NAME)
//	water X Y / pond X Y / sea edge            water squares, ponds, water off the map (game.Scenery)
//	prop TEX X Y SIZE [FRONT [SINK]]             a billboard (game.Prop)
//	well TEX X Y HALFW HALFD RIMW [FRONT]      a hole in the floor, TEX its rim
//	twosided V                                 wall value V also drawn from its open far side
func areaLooks(id, text string) []adventure.AreaLook {
	var out []adventure.AreaLook
	for _, l := range strings.Split(strings.TrimSpace(text), "\n") {
		f := strings.Fields(l)
		var set rules.Set
		var geo int
		bad := len(f) < 3
		if !bad {
			_, err := fmt.Sscan(f[0]+" "+f[1], &set, &geo)
			bad = err != nil
		}
		var a *adventure.AreaLook
		if !bad {
			if a = (&adventure.WallTextures{Areas: out}).Area(set, geo); a == nil {
				out = append(out, adventure.AreaLook{Set: set, Geo: geo})
				a = &out[len(out)-1]
			}
		}
		nums := func(from, need int, v ...*float64) bool { // the fields from from on into v, at least need of them
			if len(f) < from+need || len(f) > from+len(v) {
				return false
			}
			for i, p := range v[:len(f)-from] {
				if _, err := fmt.Sscan(f[from+i], p); err != nil {
					return false
				}
			}
			return true
		}
		var x, y float64
		switch {
		case bad:
		case f[2] == "floor" && len(f) == 4:
			a.Floor = f[3]
		case f[2] == "ceiling" && len(f) == 4:
			a.Sky = f[3]
		case f[2] == "sea":
			a.Scenery.Sea = true
		case (f[2] == "water" || f[2] == "pond") && nums(3, 2, &x, &y):
			sq := [2]int{int(x), int(y)}
			if f[2] == "water" {
				a.Scenery.Water = append(a.Scenery.Water, sq)
			} else {
				a.Scenery.Pond = append(a.Scenery.Pond, sq)
			}
		case f[2] == "prop" && len(f) >= 4:
			p := game.Prop{Tex: f[3]}
			bad = !nums(4, 3, &p.X, &p.Y, &p.Size, &p.Front, &p.Sink)
			a.Scenery.Props = append(a.Scenery.Props, p)
		case f[2] == "well" && len(f) >= 4:
			p := game.Prop{Tex: f[3], Ring: &[3]float64{}}
			bad = !nums(4, 5, &p.X, &p.Y, &p.Ring[0], &p.Ring[1], &p.Ring[2], &p.Front)
			a.Scenery.Props = append(a.Scenery.Props, p)
		case f[2] == "twosided" && nums(3, 1, &x):
			a.Scenery.TwoSided = append(a.Scenery.TwoSided, int(x))
		default:
			bad = true
		}
		if bad {
			panic(fmt.Sprintf("walls/%s_areas.txt: %q", id, l))
		}
	}
	return out
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
