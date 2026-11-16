package importer

import (
	"testing"

	"gobox/asset"
)

// TestWallTextures: every game's table parses and names pool textures.
func TestWallTextures(t *testing.T) {
	const series = "pool"
	for _, x := range wallTextures("br1", "", "").Walls { // BR1's walls are in the buck-rogers series
		if !asset.Has("buck-rogers", x.Texture) {
			t.Errorf("br1: %+v: no texture", x)
		}
	}
	for n := range solidWalls() {
		if !asset.Has(series, n) {
			t.Errorf("walls/solid.txt: no texture %s", n)
		}
	}
	for _, id := range []string{"por", "cotab", "sotsb", "pod"} {
		w := wallTextures(id, "", "")
		if len(w.Walls) == 0 {
			t.Fatalf("%s: no walls", id)
		}
		for _, x := range w.Walls {
			if !asset.Has(series, x.Texture) {
				t.Errorf("%s: %+v: no texture", id, x)
			}
		}
		for _, a := range w.Areas {
			names := []string{"floor_" + a.Floor, "ceiling_" + a.Sky}
			for _, p := range a.Scenery.Props {
				names = append(names, p.Tex)
			}
			for _, n := range names {
				if !asset.Has(series, n) {
					t.Errorf("%s: set %d geo %d: no texture %s", id, a.Set, a.Geo, n)
				}
			}
		}
	}
}
