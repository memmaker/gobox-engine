package classic

import (
	"strconv"

	"gobox/front"
	"gobox/game"
)

// Border is v with the window border's pieces from art (Data.Art):
// "border/<place>" (0..7, ui.TopLeft ..) and "border/cap/<n>" (0..3). v's
// own where art has none of it.
func Border(v front.View, art map[string]game.Picture) front.View {
	for i := range v.BorderPix {
		if p := pix(art["border/"+strconv.Itoa(i)]); p.Pix != nil {
			v.BorderPix[i] = p
		}
	}
	for i := range v.BorderCapsPix {
		if p := pix(art["border/cap/"+strconv.Itoa(i)]); p.Pix != nil {
			v.BorderCapsPix[i] = p
		}
	}
	return v
}
