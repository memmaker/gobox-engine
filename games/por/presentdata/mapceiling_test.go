package presentdata_test

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"gobox/dax"
	"gobox/geo"

	"por/presentdata"
	"por/presentdata/blobberart"
)

// TestMapCeilingInside: a remastered map's ceiling line is a sky exactly when
// the original's map is not inside (its cells' inside bit, 0x80, the one
// State's Inside reads, on all but at most one cell): -open, with no game to
// say Inside, shows the cave's rock, not clouds.
func TestMapCeilingInside(t *testing.T) {
	l, err := dax.NewLoader("../../../data/por/sources/dos")
	if err != nil {
		t.Skip("no game files")
	}
	l.UseAmiga("../../../data/por/sources/amiga")
	for g, name := range presentdata.AreaMaps {
		var b []byte
		for s := 1; s <= 8 && b == nil; s++ { // the set whose GEO file holds the map (ovr30:10C7 reads the area's set alone)
			b = l.BlockShared(fmt.Sprintf("GEO%d", s), g)
		}
		if len(b) < geo.Size {
			t.Fatalf("%s: no GEO block %d", name, g)
		}
		m, err := geo.ParseMap(b[:geo.Size])
		if err != nil {
			t.Fatal(err)
		}
		in := 0
		for y := range 16 {
			for x := range 16 {
				if m.MapCell(x, y) > 0x7F {
					in++
				}
			}
		}
		art, err := blobberart.FS.ReadFile("maps/" + name + ".map")
		if err != nil {
			t.Fatal(err)
		}
		ceil := ""
		for sc := bufio.NewScanner(strings.NewReader(string(art))); sc.Scan(); {
			if f := strings.Fields(sc.Text()); len(f) == 2 && f[0] == "ceiling" {
				ceil = f[1]
			}
		}
		if inside, sky := in >= 255, strings.HasPrefix(ceil, "sky_"); inside == sky {
			t.Errorf("%s (geo %d): %d/256 cells inside, ceiling %q", name, g, in, ceil)
		}
	}
}
