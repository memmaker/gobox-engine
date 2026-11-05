package classic

import (
	"image/color"

	"gobox/front"
	"gobox/game"
)

// Screen is v with the page's size in cells and the text colours s holds.
// v's where s has none.
func Screen(v front.View, s game.State) front.View {
	if s.Cells.X > 0 {
		v.Cells = s.Cells
	}
	v.Screen = s.Screen
	if s.Ink != ([16]color.RGBA{}) {
		v.Ink = s.Ink
	}
	return v
}

// Read is v with the journal text s.Read holds. v where s has none.
func Read(v front.View, s game.State) front.View {
	v.Read = s.Read
	return v
}
