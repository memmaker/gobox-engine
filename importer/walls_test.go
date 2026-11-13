package importer

import (
	"testing"

	"gobox/asset"
)

// TestWallTextures: every game's table parses and names pool textures.
func TestWallTextures(t *testing.T) {
	const series = "pool"
	for _, id := range []string{"cotab", "sotsb", "pod"} {
		w := wallTextures(id, "", "")
		if len(w.Walls) == 0 {
			t.Fatalf("%s: no walls", id)
		}
		for _, x := range w.Walls {
			if !asset.Has(series, x.Texture) {
				t.Errorf("%s: %+v: no texture", id, x)
			}
		}
	}
}
