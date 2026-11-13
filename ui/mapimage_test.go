package ui

import (
	"image"
	"image/color"
	"testing"
)

func TestMapImage(t *testing.T) {
	var ink [16]color.RGBA
	for i := range ink {
		ink[i] = color.RGBA{uint8(i*16 + 1), uint8(i), 7, 255}
	}
	door := color.RGBA{1, 2, 3, 255}
	doorRGBA, labelInks = door, []int{10}
	m := &MapSquares{Party: image.Pt(1, 1), Ahead: image.Pt(0, -1)}
	m.Squares[0][0] = MapSquare{Seen: Visited, Walls: [4]MapWall{Wall, Door, NoWall, NoWall}}
	m.Squares[0][2] = MapSquare{Seen: SeenFar, Kind: 0, Letter: 'A'}
	img := mapImage(m, 8, ink, nil)
	for _, c := range []struct {
		p image.Point
		w color.RGBA
	}{
		{image.Pt(3, 3), ink[1]},    // visited
		{image.Pt(3, 0), ink[15]},   // north wall
		{image.Pt(7, 3), door},      // east door
		{image.Pt(17, 1), ink[8]},   // seen far
		{image.Pt(20, 3), ink[10]},  // label dot (cell 8: 2..5)
		{image.Pt(12, 12), ink[12]}, // party
		{image.Pt(100, 100), ink[0]},
	} {
		if got := img.RGBAAt(c.p.X, c.p.Y); got != c.w {
			t.Errorf("%v = %v, want %v", c.p, got, c.w)
		}
	}
}
