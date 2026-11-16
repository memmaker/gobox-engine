package engine

import "testing"

// TestBR2WallsBackdrop: BR2 on the shared engine (BACK1 only, flats buckrogers2-go rules/gamewalls.go BackFlats, tested as
// BR1's TestBR1WallsBackdrop): the base's streets without a BACK picture show the city's flats; with BACK1 block 17
// loaded last a square with bit 7 (indoors) shows that picture and the haze's flats.
func TestBR2WallsBackdrop(t *testing.T) {
	e := brBegun(t, gameBR2)
	w := br1Walls(t, e)
	if w.Floor != "br1_city" || w.Sky != "sky_br1_city" || w.Backdrop != nil {
		t.Errorf("no picture: floor %q sky %q backdrop %v", w.Floor, w.Sky, w.Backdrop != nil)
	}
	e.LoadWalldef(1, 1)
	e.LoadWalldef(17, 2)
	e.gs.geo.Plane[0x200+e.posY<<4+e.posX] |= 0x80
	w = br1Walls(t, e)
	if w.Floor != "br1_mars_haze" || w.Sky != "sky_br1_mars_haze" || w.Backdrop == nil || w.Backdrop.Block != 17 || w.Backdrop.File != "BACK1" {
		t.Errorf("indoors: floor %q sky %q backdrop %+v", w.Floor, w.Sky, w.Backdrop)
	}
	e.gs.geo.Plane[0x200+e.posY<<4+e.posX] &^= 0x80
	w = br1Walls(t, e)
	if w.Floor != "br1_city" || w.Backdrop != nil { // 17+128 is no BACK1 block
		t.Errorf("outdoors: floor %q backdrop %v", w.Floor, w.Backdrop != nil)
	}
}
