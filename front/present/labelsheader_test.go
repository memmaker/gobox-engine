package present

import (
	"image"
	"image/color"
	"testing"

	"gobox/front"
)

// TestFullMapNoLabelsHeader: a map with no labels draws no "Labels" header.
func TestFullMapNoLabelsHeader(t *testing.T) {
	g, _ := playGame(mapState)
	p := &Presenter{}
	p.SetEngine(g)
	p.fullMap = &fullMapUI{}
	p.shownView = front.View{Cells: image.Pt(40, 25)}
	m := image.NewRGBA(image.Rect(0, 0, 320, 200))
	full := func(byte, bool) [8]byte { return [8]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF} }
	p.drawModal(m, p.shownView, Note{Font: full, Glyph: func(byte) [8]byte { return [8]byte{} }})
	fm := p.fm()
	y := (fm.LegendAt.Y - 1) * 8
	for x := fm.LegendAt.X * 8; x < m.Rect.Dx(); x++ {
		if c := m.RGBAAt(x, y+3); c != (color.RGBA{A: 255}) && c != (color.RGBA{}) {
			t.Fatalf("a Labels header at (%d,%d) with no labels: %v", x, y+3, c)
		}
	}
}
