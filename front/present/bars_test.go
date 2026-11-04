package present

import "testing"

// Without Data.BarsPNG the cards' bars are drawn in the theme's colours.
func TestBarArtWithoutPNG(t *testing.T) {
	old := barsPNG
	defer func() { barsPNG = old }()
	barsPNG = nil
	m := barArt()
	if m.Bounds() != goldFill.Union(barRight) || m.At(greenFill.Min.X, greenFill.Min.Y) != green {
		t.Fatalf("bar art %v", m.Bounds())
	}
}
