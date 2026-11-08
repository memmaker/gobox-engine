package classic

import (
	"image"
	"image/color"

	"gobox/front"
	"gobox/game"
)

// Screen is v with the typed screen up and the page's size in cells and text
// colours; v's where those are unset.
func Screen(v front.View, cells image.Point, sc *game.Screen, ink [16]color.RGBA) front.View {
	if cells.X > 0 {
		v.Cells = cells
	}
	v.Screen = sc
	if ink != ([16]color.RGBA{}) {
		v.Ink = ink
	}
	return v
}
