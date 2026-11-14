package importer

import (
	"fmt"
	"strings"
	"testing"

	"gobox/asset/orig"
	"gobox/rules"
)

// TestAreaSkyInside: a POR area's own ceiling (walls/por_areas.txt) is a sky
// exactly when the original's map is not inside (its cells' inside bit,
// 0x80, the one State's Inside reads, on all but at most one cell).
func TestAreaSkyInside(t *testing.T) {
	l, err := orig.NewLoader("../data/por/sources/dos")
	if err != nil {
		t.Skip("no game files")
	}
	for _, a := range wallTextures("por", "", "").Areas {
		b := l.BlockShared(a.Set.File("GEO"), a.Geo)
		if len(b) < rules.GeoSize {
			t.Fatalf("set %d: no GEO block %d", a.Set, a.Geo)
		}
		m, err := rules.ParseMap(b[:rules.GeoSize])
		if err != nil {
			t.Fatal(err)
		}
		in := 0
		for y := range rules.AreaSize {
			for x := range rules.AreaSize {
				if m.MapCell(x, y) > 0x7F {
					in++
				}
			}
		}
		if inside, sky := in >= rules.AreaSize*rules.AreaSize-1, strings.HasPrefix(a.Sky, "sky_"); inside == sky {
			t.Errorf("%s: %d cells inside, ceiling %q", fmt.Sprint(a.Set, "/", a.Geo), in, a.Sky)
		}
	}
}
