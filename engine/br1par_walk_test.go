package engine

import (
	"testing"

	"gobox/game"
)

// arriveOffered is whether the state offers the step's Arrive.
func arriveOffered(e *Engine) bool {
	for _, a := range commonOf(e.build()).Allowed.Plain {
		if _, ok := a.(game.Arrive); ok {
			return true
		}
	}
	return false
}

// TestBR1WallBump ports rules TestClassWallBump (bug class 87): a step into a wall is RunEvent{Before} alone, no
// Arrive is offered; the walk ends there and the next key acts: a turn to an open side, then its step offers Arrive.
func TestBR1WallBump(t *testing.T) {
	e := br1Begun(t)
	wall, open := -1, -1
	for i := 0; i < 256 && (wall < 0 || open < 0); i++ {
		e.posX, e.posY, wall, open = i%16, i/16, -1, -1
		for s := 0; s < 8; s += 2 {
			if e.DoorAt(e.posX, e.posY, s) == 0 {
				wall = s
			} else {
				open = s
			}
		}
	}
	if wall < 0 || open < 0 {
		t.Fatalf("no square with a wall and an open side (%d,%d)", e.posX, e.posY)
	}
	x, y := e.posX, e.posY
	if _, err := e.doNow(game.RunEvent{Before: true, Side: wall}); err != nil {
		t.Fatal(err)
	}
	if arriveOffered(e) || e.posX != x || e.posY != y {
		t.Fatalf("a step into the wall: Arrive offered %v, at %d,%d from %d,%d", arriveOffered(e), e.posX, e.posY, x, y)
	}
	if _, err := e.doNow(game.Face{Facing: open}); err != nil || e.facing != open {
		t.Fatalf("turn to %d: facing %d, %v", open, e.facing, err)
	}
	if _, err := e.doNow(game.RunEvent{Before: true, Side: open}); err != nil {
		t.Fatal(err)
	}
	for range 10 {
		if commonOf(e.build()).Ask == nil {
			break
		}
		press(e)
	}
	if e.posX != x || e.posY != y || !arriveOffered(e) {
		t.Errorf("a step into the open side: Arrive offered %v at %d,%d", arriveOffered(e), e.posX, e.posY)
	}
}
