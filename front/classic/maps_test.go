package classic

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/game"
)

func TestMap(t *testing.T) {
	s := game.State{Where: game.Where{X: 2, Y: 3, Facing: 2}, Map: game.Automap{Size: image.Pt(16, 16),
		Walls: []game.Wall{{At: image.Pt(1, 1), Side: 6}}, Doors: []game.Wall{{At: image.Pt(1, 1), Side: 0, Kind: 2}},
		Marked: []image.Point{{4, 4}, {5, 5}}, Notes: []string{"inn", "Bank"}}}
	m := Map(&front.MapSquares{}, s)
	if q := m.Squares[1][1]; q.Walls != [4]front.MapWall{front.LockedDoor, 0, 0, front.Wall} || m.Party != image.Pt(2, 3) || m.Ahead != image.Pt(1, 0) {
		t.Errorf("%+v %v %v", q, m.Party, m.Ahead)
	}
	if m.Squares[4][4].Letter != 'B' || m.Squares[5][5].Letter != 'A' {
		t.Errorf("letters %c %c", m.Squares[4][4].Letter, m.Squares[5][5].Letter)
	}
	if f := FullMap(&front.FullMap{Map: m, Cursor: image.Pt(4, 4)}, s); f.Info != "4,4 inn" || !f.Labelled {
		t.Errorf("full map %+v", f)
	}
}
