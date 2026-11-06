package grid

import "testing"

// Pinned on the games' walkers: a straight line costs 2 half squares a
// step, a diagonal 3; the quarter arcs and sectors as the games had them.
func TestGridPinned(t *testing.T) {
	l := NewLine(0, 0, 4, 0)
	n := 0
	for l.Step() {
		n++
	}
	if n != 4 || l.Steps != 8 || l.Dir != 8 {
		t.Errorf("straight: %d steps, %d half squares, dir %d", n, l.Steps, l.Dir)
	}
	d := NewLine(0, 0, 3, 3)
	d.Step()
	if d.Steps != 3 || d.Dir != 3 {
		t.Errorf("diagonal: %d half squares, dir %d", d.Steps, d.Dir)
	}
	if !InArc(50, 25, 5, 5, 5, 3, 0) || InArc(50, 25, 5, 5, 5, 8, 0) || InArc(50, 25, -1, 5, 5, 3, 8) {
		t.Error("InArc")
	}
	if DirectionTo(0, 0, 10, 0, 618, 106, 256, true) != 2 || DirectionTo(0, 0, 0, 5, 618, 106, 256, true) != 4 || DirectionTo(0, 0, 4, 4, 618, 106, 256, true) != 3 {
		t.Error("DirectionTo")
	}
	if !FormationEdge(-1, -1, 11, 6) || FormationEdge(3, -1, 11, 6) {
		t.Error("FormationEdge")
	}
	rng, tx, ty := 10, 6, 0
	open := Sight{Height: func(x, y int) (int, int) { return 1, 0 }}
	if !LineOfSight(open, &rng, &tx, &ty, 0, 0) || rng != 12 {
		t.Errorf("open sight: rng %d", rng)
	}
	wall := Sight{Height: func(x, y int) (int, int) {
		if x == 3 {
			return 1, 9
		}
		return 1, 0
	}}
	rng, tx, ty = 10, 6, 0
	if LineOfSight(wall, &rng, &tx, &ty, 0, 0) || tx != 3 {
		t.Errorf("wall: stopped at %d", tx)
	}
	type e struct{ dist, dir byte }
	l2 := []e{{5, 1}, {2, 3}, {2, 1}}
	NearSort(l2, func(x e) (byte, byte) { return x.dist, x.dir }, true)
	if l2[0] != (e{2, 1}) {
		t.Errorf("NearSort low first: %v", l2)
	}
}
