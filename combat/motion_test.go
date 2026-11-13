package combat

import (
	"image"
	"math"
	"testing"
)

// A scripted fight's boards as the game hands them over: a walk the board
// offered a path for, a monster's whole move in one board, a fall, a
// newcomer. Frame by frame no figure moves more than a step (no teleport),
// every walk arrives, the fallen sink away and the newcomer rises.
func TestNoTeleport(t *testing.T) {
	icon := image.NewRGBA(image.Rect(0, 0, 24, 24))
	unit := func(id, x, y int) Unit {
		return Unit{ID: id, X: x, Y: y, Size: 1, Squares: []image.Point{{x, y}}, Icon: icon}
	}
	path := map[image.Point][]Step{{5, 6}: {{Point: image.Pt(2, 2)}, {Point: image.Pt(3, 3)}, {Point: image.Pt(4, 4)}, {Point: image.Pt(5, 5)}, {Point: image.Pt(5, 6)}}}
	boards := []*Board{
		{Units: []Unit{unit(1, 2, 2), unit(2, 10, 10)}, Paths: path},
		{Units: []Unit{unit(1, 5, 6), unit(2, 10, 10)}}, // walked the offered path
		{Units: []Unit{unit(1, 5, 6), unit(2, 3, 10)}},  // a whole move in one board
		{Units: []Unit{unit(1, 5, 6)}},                  // 2 fell
		{Units: []Unit{unit(1, 5, 6), unit(3, 8, 8)}},   // 3 appeared
	}
	const dt = 1.0 / 60
	var p CombatPresenter
	last := map[int][2]float32{}
	for i, b := range boards {
		for f := 0; f < 3*60; f++ {
			for _, s := range p.moveUnits(b, dt) {
				if was, ok := last[s.u.ID]; ok {
					if d := math.Max(math.Abs(float64(s.x-was[0])), math.Abs(float64(s.z-was[1]))); d > maxStepsAFrame || d > 0.25 {
						t.Fatalf("board %d frame %d: unit %d jumped %.2f squares", i, f, s.u.ID, d)
					}
				}
				last[s.u.ID] = [2]float32{s.x, s.z}
				if i == 4 && s.u.ID == 3 && f == 0 && s.scale >= 1 {
					t.Error("the newcomer popped up")
				}
			}
		}
		for _, u := range b.Units {
			if w := p.units[u.ID]; w.at != unitCentre(u) || w.grow != 1 {
				t.Errorf("board %d: unit %d at %v, not on its square", i, u.ID, w.at)
			}
		}
		if i == 1 && p.units[1].at != unitCentre(b.Units[0]) {
			t.Error("walk did not arrive")
		}
	}
	if _, ok := p.units[2]; ok {
		t.Error("the fallen never left")
	}
}

// A walk takes a steady step time per square: the walk of path squares
// (no catch-up) arrives after that many steps, not at once.
func TestStepPace(t *testing.T) {
	icon := image.NewRGBA(image.Rect(0, 0, 24, 24))
	p := CombatPresenter{StepTime: 0.2}
	p.moveUnits(&Board{Units: []Unit{{ID: 1, X: 0, Y: 0, Squares: []image.Point{{0, 0}}, Icon: icon}}}, 0.01)
	b := &Board{Units: []Unit{{ID: 1, X: 2, Y: 0, Squares: []image.Point{{2, 0}}, Icon: icon}}}
	p.moveUnits(b, 0.2) // one step
	if w := p.units[1]; w.at != [2]float64{1.5, 0.5} {
		t.Errorf("after one step time at %v", w.at)
	}
}
