package classic

import (
	"strconv"

	"gobox/front"
	"gobox/game"
)

// Place is v with the view window and the party's place from s: the window
// (numbered as front.Window), the area, the square and the way it faces.
func Place(v front.View, s game.State) front.View {
	a, _ := strconv.Atoi(s.Where.Area)
	v.Window, v.Area = front.Window(s.Window), front.AreaID(a)
	v.X, v.Y, v.Facing = s.Where.X, s.Where.Y, front.Dir(s.Where.Facing/2&3)
	return v
}
