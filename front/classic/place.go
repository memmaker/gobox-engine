package classic

import (
	"strconv"

	"gobox/front"
	"gobox/game"
)

// Place is v with the view window and the party's place from where: the window
// (numbered as front.Window), the area, the square and the way it faces.
func Place(v front.View, where game.Where, window game.Window) front.View {
	a, _ := strconv.Atoi(where.Area)
	v.Window, v.Area = front.Window(window), front.AreaID(a)
	v.X, v.Y, v.Facing = where.X, where.Y, front.Dir(where.Facing/2&3)
	if q := where.Square; q != nil {
		v.X, v.Y = q.X, q.Y
	}
	return v
}
