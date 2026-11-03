package blobber

import (
	"image"
	"testing"
)

// The host's Textures replace the embedded walls; nil keeps them; the sky reads the current View.
func TestViewTexturesAndSky(t *testing.T) {
	defer func(c *View) { cur = c }(cur)
	cur = &View{}
	if decodeWall("stone") == nil {
		t.Fatal("nil Textures: the embedded stone should load")
	}
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	cur = &View{Textures: map[string]image.Image{"wall1": img}}
	if decodeWall("wall1") != img || decodeWall("stone") != nil {
		t.Fatal("host Textures: only the host's images")
	}
	if night() {
		t.Fatal("day by default")
	}
	cur.Night = true
	if !night() {
		t.Fatal("View.Night")
	}
}

// sunAt-like bearings wrap into -pi..pi.
func TestWrap(t *testing.T) {
	for _, c := range [][2]float64{{0, 0}, {7, 7 - 6.283185307179586}, {-4, -4 + 6.283185307179586}} {
		if d := wrap(c[0]) - c[1]; d > 1e-9 || d < -1e-9 {
			t.Fatalf("wrap(%v) = %v", c[0], wrap(c[0]))
		}
	}
}
