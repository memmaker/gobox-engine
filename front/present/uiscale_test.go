package present

import (
	"testing"

	"gobox/gpu"
)

// TestUIScaleFits: whatever the Portraits and Minimap sizes, the eight cards
// and the map stay on screen and apart.
func TestUIScaleFits(t *testing.T) {
	defer func(o Options) { *Host = o }(*Host)
	for _, sz := range [][2]int{{640, 400}, {1904, 1204}, {3840, 2160}} {
		dst := gpu.NewImage(sz[0], sz[1])
		for _, c := range UISteps {
			for _, m := range UISteps {
				for _, row := range []bool{false, true} {
					Host.PartyRow = row
					Host.Cards, Host.Map = c, m
					mr, at := mapRect(dst, 8), cardAt(dst)
					for i := range 8 {
						r := at(i, 8)
						if !r.In(dst.Bounds()) || r.Overlaps(mr) || i > 0 && (row || i != 4) && r.Overlaps(at(i-1, 8)) {
							t.Fatalf("%v cards %v map %v row %v: card %d %v, map %v", sz, c, m, row, i, r, mr)
						}
					}
				}
			}
		}
		dst.Deallocate()
	}
}
