package pool

import (
	"rules1e"
	"testing"
)

// TestAIPathAroundWall: a wall stands between two foes, open only at its
// top end; the first step heads up towards the gap, not straight at the
// wall.
func TestAIPathAroundWall(t *testing.T) {
	e := &Engine{}
	e.initVM()
	e.state().Cmap = &rules1e.CombatMap{} // a fight's map (BuildGround makes it)
	for y := range e.state().Cmap.Ground {
		for x := range e.state().Cmap.Ground[y] {
			e.state().Cmap.Ground[y][x] = rules1e.GroundFloor
		}
	}
	for y := 3; y < 25; y++ {
		e.state().Cmap.Ground[y][10] = 1 // wall at x 10, gap above y 3
	}
	a, b := &rules1e.Character{}, &rules1e.Character{}
	e.state().CombatantSlots, e.state().Ncomb = []*rules1e.Character{a, b}, 2
	e.state().Cpos[1] = rules1e.CombatPos{X: 8, Y: 10, Size: 1}
	e.state().Cpos[2] = rules1e.CombatPos{X: 12, Y: 10, Size: 1}
	e.state().Occ[10][8], e.state().Occ[10][12] = 1, 2
	d := e.state().AiPathDir(a, b)
	if d < 0 || rules1e.DirDY[d] >= 0 {
		t.Fatalf("first step %d (dx %d dy %d) does not head for the gap", d, rules1e.DirDX[d], rules1e.DirDY[d])
	}
}
