package playtest

import (
	"gobox/game/play"
	"testing"
)

// TestDeskewPlacement: on a deskewed map each side stands on its own
// cell's floor, the monsters in the cell ahead of the party, so nothing but
// the open cell edge lies between them (the start of Phlan: a corridor
// facing west).
func TestDeskewPlacement(t *testing.T) {
	e := explore(t)
	e.Deskew.Store(true)
	e.DevAction = e.DevScene("combat")
	seed := uint32(1)
	e.FixedSeed = &seed
	e.Init()
	e.Start(e.Play)
	v, ok := settle(e).(play.Fighting)
	for range 7 {
		if ok {
			break
		}
		pressKey(t, e)
		v, ok = settle(e).(play.Fighting)
	}
	if !ok {
		t.Fatal("no fight")
	}
	for _, u := range v.Field.Units {
		x0 := 22 // the party's cell: floor columns 22..26, rows 11..14
		if u.Foe {
			x0 -= 6 // the cell west of it
		}
		if u.At.X < x0 || u.At.X > x0+4 || u.At.Y < 11 || u.At.Y > 14 {
			t.Errorf("%s (party %v) at %d,%d, want x %d..%d, y 11..14", u.Name, !u.Foe, u.At.X, u.At.Y, x0, x0+4)
		}
	}
}
