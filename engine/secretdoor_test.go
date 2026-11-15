package engine

import (
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestSecretDoorOnMap: the Hobgoblin Barracks' secret door (13,2 east: a
// door behind a wall that looks solid) is a wall on the automap until the
// party has gone through it, then a door from both its cells; a door that
// looks like one is a door from the start. Kept with the save.
func TestSecretDoorOnMap(t *testing.T) {
	e := testEngine(t)
	if e.game().Walls == nil {
		t.Skip("no POR wall table imported")
	}
	e.viewMode, e.set = 1, 4
	e.mapPairA = [rules.WallSlots]int{-1, -1, -1}
	e.LoadGeo(2)
	for slot, block := range []int{1, 3, 20} { // LoadWalls(Stone, Overgrown, Ruins)
		e.LoadWalldef(block, slot+1)
	}
	door := func(x, y, side int) bool {
		at := func(w game.Wall) bool { return w.At == (rules.MapPos{X: x, Y: y}) && w.Side == side }
		m := e.mapState(&game.Where{})
		isDoor, isWall := slices.ContainsFunc(m.Doors, at), slices.ContainsFunc(m.Walls, at)
		if isDoor == isWall {
			t.Fatalf("(%d,%d) side %d: door %v, wall %v", x, y, side, isDoor, isWall)
		}
		return isDoor
	}
	if door(13, 2, 2) || door(14, 2, 6) {
		t.Error("the secret door shows before it is found")
	}
	if !door(13, 0, 2) {
		t.Error("a plain door (13,0 east) does not show")
	}
	e.posX, e.posY, e.facing = 13, 2, 2
	e.StepForward()
	if !door(13, 2, 2) || !door(14, 2, 6) {
		t.Error("the secret door gone through does not show")
	}
	found := e.ext.secrets
	if err := e.saveTo("A"); err != nil {
		t.Fatal(err)
	}
	e.ext.secrets = nil
	if !e.loadFrom("A") || len(e.ext.secrets[e.areaKey()]) != len(found[e.areaKey()]) {
		t.Errorf("the found secret doors are not kept with the save: %v, want %v", e.ext.secrets, found)
	}
}
