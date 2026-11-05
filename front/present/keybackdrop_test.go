package present

import (
	"image"
	"image/color"
	"testing"
)

// A DAX sprite's flat backdrop goes transparent where it reaches the edge;
// the same colour inside the figure stays, and a keyed image is reused.
func TestKeyBackdrop(t *testing.T) {
	bg, fig := color.RGBA{82, 82, 82, 255}, color.RGBA{200, 0, 0, 255}
	m := image.NewRGBA(image.Rect(0, 0, 5, 5))
	for i := 0; i < len(m.Pix); i += 4 {
		copy(m.Pix[i:], []byte{bg.R, bg.G, bg.B, 255})
	}
	for _, p := range [][2]int{{1, 1}, {2, 1}, {3, 1}, {1, 2}, {3, 2}, {1, 3}, {2, 3}, {3, 3}} {
		m.SetRGBA(p[0], p[1], fig)
	}
	k := keyBackdrop(m).(*image.RGBA)
	if k.RGBAAt(0, 0).A != 0 || k.RGBAAt(4, 4).A != 0 {
		t.Error("edge backdrop kept")
	}
	if k.RGBAAt(2, 2) != bg || k.RGBAAt(1, 1) != fig {
		t.Error("inside the figure changed")
	}
	if keyBackdrop(m) != image.Image(k) || keyBackdrop(k) != image.Image(k) {
		t.Error("not stable")
	}
}
