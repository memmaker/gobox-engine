package present

import (
	"image"
	"testing"

	"gobox/front"
)

// TestPicPlaceOnScreen: the picture box (the view window's whole for
// WinText, the inner frame for a picture) lies inside the space it is
// placed in at every window size: it never runs off the top (Rolf's tour
// drew WinText's tall box placed by the inner frame's height).
func TestPicPlaceOnScreen(t *testing.T) {
	txt := paneCells(&front.Pane{Lines: make([][]front.PaneRun, 6)}).Framed()
	th := float64(txt.Pixels().Dy())
	for _, pic := range []front.FrontCellRect{innerWindow.Framed(), viewWindow.Framed()} {
		ph := float64(pic.Pixels().Dy())
		for _, sz := range []image.Point{{400, 250}, {640, 400}, {1280, 800}, {1920, 1080}, {2560, 1600}, {1000, 330}} {
			b := image.Rect(0, 0, sz.X, sz.Y)
			for _, k := range []float64{1, 2, 3, 4} {
				for _, hp := range []float64{1, 2, 3, 4} { // Graphics' picture size
					old := Host.Pic
					Host.Pic = hp
					pk := min(k, max(1, float64(min(int(float64(b.Dy())/2/(ph+th/2)), b.Dx()/8))))
					pp, py, _ := picPlace(pic, pk, 1, 1, th, b)
					if top, bot := float64(py)-pp*ph/2, float64(py)+pp*ph/2; top < 0 || bot > float64(b.Dy()) {
						t.Errorf("box %v in %v at k %v pic %v: rows %.0f..%.0f off the screen", pic, sz, k, hp, top, bot)
					}
					Host.Pic = old
				}
			}
		}
	}
}
