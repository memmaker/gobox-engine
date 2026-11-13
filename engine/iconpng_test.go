package engine

import (
	"image"
	"image/color"
	"gobox/rules1e"
	"testing"
)

// A square PNG is both poses; one twice as wide is cut in the middle, the
// left half ready, the right action, mirrored to face left; other shapes
// and bytes that are no PNG are refused.
func TestIconPNG(t *testing.T) {
	wide := image.NewRGBA(image.Rect(0, 0, 4, 2))
	wide.Set(0, 0, color.White) // ready: top left
	wide.Set(3, 1, color.White) // action: bottom right
	e := &Engine{}
	var c rules1e.Character
	if e.takeIconPNG(e.charID(&c), pngOf(t, image.NewRGBA(image.Rect(0, 0, 3, 1)))) || c.Icon != nil {
		t.Fatal("3:1 taken")
	}
	if e.takeIconPNG(e.charID(&c), []byte("no png")) || c.Icon != nil {
		t.Fatal("not a PNG taken")
	}
	if !e.takeIconPNG(e.charID(&c), pngOf(t, wide)) {
		t.Fatal("2:1 refused")
	}
	at := func(im image.Image, x, y int) bool { _, _, _, a := im.At(x, y).RGBA(); return a != 0 }
	if r := e.iconPNG(e.charID(&c), 0, false); r.Bounds().Dx() != 2 || !at(r, 0, 0) || at(r, 1, 1) {
		t.Fatal("ready is not the left half")
	}
	if a := e.iconPNG(e.charID(&c), 1, false); !at(a, 1, 1) || at(a, 0, 0) {
		t.Fatal("action is not the right half")
	}
	if a := e.iconPNG(e.charID(&c), 1, true); !at(a, 0, 1) {
		t.Fatal("the flip did not mirror")
	}
	if !e.takeIconPNG(e.charID(&c), pngOf(t, image.NewRGBA(image.Rect(0, 0, 2, 2)))) || e.iconPNG(e.charID(&c), 0, false) != e.iconPNG(e.charID(&c), 1, false) {
		t.Fatal("a square is not both poses, cached")
	}
}
