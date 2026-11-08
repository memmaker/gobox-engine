package classic

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/game"
)

func TestMap(t *testing.T) {
	s := shown{Where: game.Where{X: 2, Y: 3, Facing: 2}, Map: game.Automap{Size: image.Pt(16, 16),
		Walls: []game.Wall{{At: image.Pt(1, 1), Side: 6}}, Doors: []game.Wall{{At: image.Pt(1, 1), Side: 0, State: game.DoorLocked}},
		Marked: []image.Point{{4, 4}, {5, 5}}, Notes: []string{"inn", "Bank"}}}
	m := Map(&front.MapSquares{}, s.Map, s.Where)
	if q := m.Squares[1][1]; q.Walls != [4]front.MapWall{front.LockedDoor, 0, 0, front.Wall} || m.Party != image.Pt(2, 3) || m.Ahead != image.Pt(1, 0) {
		t.Errorf("%+v %v %v", q, m.Party, m.Ahead)
	}
	if m.Squares[4][4].Letter != 'B' || m.Squares[5][5].Letter != 'A' {
		t.Errorf("letters %c %c", m.Squares[4][4].Letter, m.Squares[5][5].Letter)
	}
	if f := FullMap(&front.FullMap{Map: m, Cursor: image.Pt(4, 4)}, s.Map, s.Where); f.Info != "4,4 inn" || !f.Labelled {
		t.Errorf("full map %+v", f)
	}
}

// TestFullMapNoSquares: the front's M asks FullMap with no squares yet; a
// 16x16 map up must still come back with squares (nil crashed present.mapImage).
func TestFullMapNoSquares(t *testing.T) {
	s := shown{Map: game.Automap{Size: image.Pt(16, 16)}}
	if f := FullMap(&front.FullMap{}, s.Map, s.Where); f == nil || f.Map == nil {
		t.Fatalf("FullMap: %+v", f)
	}
}
