package engine

import (
	"testing"

	"gobox/game"
)

func br1Walls(t *testing.T, e *Engine) game.MapWalls {
	t.Helper()
	x, ok := e.build().(game.Exploring)
	if !ok {
		t.Fatalf("%T", e.build())
	}
	return x.Map.Walls
}

// TestBR1WallsBackdrop: the tarmac (block 16, set 1, walls 7 then 127: no BACK1 picture 255) shows the
// city's flats without a picture, as the old front did (buckrogers-go rules TestClassStreetOutside); in
// set 4 with BACK4 block 10 loaded last, outdoors (cell without bit 7: block 138) the picture and the mars
// clouds' flats show (rules TestBackdrop, TestBackFlats), indoors (bit 7: block 10) the haze's.
func TestBR1WallsBackdrop(t *testing.T) {
	e := br1Begun(t)
	if w := br1Walls(t, e); w.Floor != "br1_city" || w.Sky != "sky_br1_city" || w.Backdrop != nil {
		t.Errorf("tarmac: floor %q sky %q backdrop %v", w.Floor, w.Sky, w.Backdrop != nil)
	}
	e.set = 4
	e.LoadWalldef(4, 1)
	e.LoadWalldef(10, 2)
	w := br1Walls(t, e)
	if w.Floor != "br1_mars_clouds" || w.Sky != "sky_br1_mars_clouds" || w.Backdrop == nil || w.Backdrop.Block != 138 || w.Backdrop.File != "BACK4" {
		t.Errorf("set 4 outdoors: floor %q sky %q backdrop %+v", w.Floor, w.Sky, w.Backdrop)
	}
	e.gs.geo.Plane[0x200+e.posY<<4+e.posX] |= 0x80
	w = br1Walls(t, e)
	if w.Floor != "br1_mars_haze" || w.Backdrop == nil || w.Backdrop.Block != 10 {
		t.Errorf("set 4 indoors: floor %q backdrop %+v", w.Floor, w.Backdrop)
	}
}
