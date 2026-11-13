package engine

import (
	"gobox/rules1e"
	"image"
	"testing"
)

// A combat icon's PNG is square or twice as wide as high (ui cuts the
// poses: ownpng.go); other shapes and bytes that are no PNG are refused.
func TestIconPNG(t *testing.T) {
	e := &Engine{}
	var c rules1e.Character
	if e.takeIconPNG(e.charID(&c), pngOf(t, image.NewRGBA(image.Rect(0, 0, 3, 1)))) || c.Icon != nil {
		t.Fatal("3:1 taken")
	}
	if e.takeIconPNG(e.charID(&c), []byte("no png")) || c.Icon != nil {
		t.Fatal("not a PNG taken")
	}
	if !e.takeIconPNG(e.charID(&c), pngOf(t, image.NewRGBA(image.Rect(0, 0, 4, 2)))) || c.Icon == nil {
		t.Fatal("2:1 refused")
	}
	if !e.takeIconPNG(e.charID(&c), pngOf(t, image.NewRGBA(image.Rect(0, 0, 2, 2)))) {
		t.Fatal("a square refused")
	}
}
