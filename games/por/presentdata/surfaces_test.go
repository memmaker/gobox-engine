package presentdata

import (
	"bufio"
	"strings"
	"testing"

	"gobox/walls"
	"por/presentdata/blobberart"
)

// TestSurfacesTextured: every AreaMaps map names a floor, a ceiling or sky and
// a texture for every face and prop, and each one is in the pool walls series
// (BUG-CLASSES 12: no solid-coloured surface).
func TestSurfacesTextured(t *testing.T) {
	for geo, name := range AreaMaps {
		f, err := blobberart.FS.Open("maps/" + name + ".map")
		if err != nil {
			t.Errorf("geo %d: %v", geo, err)
			continue
		}
		var floor, ceil bool
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			w := strings.Fields(sc.Text())
			if len(w) < 2 {
				continue
			}
			tex := ""
			switch w[0] {
			case "floor":
				floor, tex = true, "floor_"+w[1]
			case "ceiling":
				ceil, tex = true, "ceiling_"+w[1]
			case "face":
				tex = w[len(w)-1]
			case "prop":
				tex = w[1]
			}
			if tex != "" && !walls.Has("pool", tex) {
				t.Errorf("%s: %q has no texture %s", name, sc.Text(), tex)
			}
		}
		f.Close()
		if !floor || !ceil {
			t.Errorf("%s: floor %v ceiling %v", name, floor, ceil)
		}
	}
}
