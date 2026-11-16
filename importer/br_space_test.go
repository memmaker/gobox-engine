package importer

import (
	"testing"

	"gobox/adventure"
	"gobox/rules"
)

// TestBR1SpaceMap: the system map in BR1's data folder is the Pascal's (gdconsts.pas orbitPositions and the tables after it,
// explore.pas, orbitmap.pas, gameinit.pas): 58 body squares, five orbits, eight belt bodies, the ship's start on orbit 2's body at
// phase 0, and the bodies at the buckrogers-go oracle's squares (adv_test.go TestRocketStep: planet 2 on (9,8) at step 0).
func TestBR1SpaceMap(t *testing.T) {
	tb, err := adventure.ReadGameTables(br1EngineFolder(t))
	if err != nil {
		t.Fatal(err)
	}
	s := tb.Space
	if s == nil {
		t.Fatal("no system map")
	}
	if len(s.Squares) != 58 || s.Squares[0] != (adventure.SpaceSquare{X: 12, Y: 10, Sprite: 1}) || s.Squares[57] != (adventure.SpaceSquare{X: 10, Y: 0, Sprite: 67}) {
		t.Errorf("squares %v", s.Squares)
	}
	if want := []adventure.SpaceOrbit{{0, 2, 0}, {2, 4, 0}, {6, 8, 7}, {14, 16, 9}, {30, 28, 7}}; len(s.Orbits) != 5 || s.Orbits[0] != want[0] || s.Orbits[1] != want[1] ||
		s.Orbits[2] != want[2] || s.Orbits[3] != want[3] || s.Orbits[4] != want[4] {
		t.Errorf("orbits %v", s.Orbits)
	}
	if len(s.Offsets) != 8 || s.Offsets[0] != 3 || s.Offsets[7] != 26 {
		t.Errorf("belt %v", s.Offsets)
	}
	if len(s.Steps) != 8 || s.Steps[0] != (rules.MapPos{X: 0, Y: -1}) || s.Steps[1] != (rules.MapPos{X: -1, Y: -1}) || s.Steps[6] != (rules.MapPos{X: 1, Y: 0}) {
		t.Errorf("steps %v", s.Steps)
	}
	if got := s.Squares[s.Start]; got.X != 9 || got.Y != 8 {
		t.Errorf("start square %v", got)
	}
	if s.Bodies() != 13 || s.Body(2, 0).At != (rules.MapPos{X: 9, Y: 8}) || s.Body(2, 0).Tile != 10+296 {
		t.Errorf("planet 2 at 0: %+v", s.Body(2, 0))
	}
	if b := s.Body(5, 0); b.At != (rules.MapPos{X: s.Squares[30+(7+3)%28].X, Y: s.Squares[30+(7+3)%28].Y}) {
		t.Errorf("belt body 5 at 0: %+v", b)
	}
	if tb2, err := adventure.ReadGameTables(brEngineFolder(t, "br2", BR2EXE)); err != nil || tb2.Space != nil {
		t.Errorf("br2 has no system map yet: %v %v", tb2.Space, err)
	}
}
