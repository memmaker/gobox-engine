package ui

import (
	"image"
	"image/color"
	"testing"
)

// TestGroundUnderProps: with no floor picture and no floor colour, the
// floor every square shows (under props and models too) is the area's own
// ground tile, never a black hole.
func TestGroundUnderProps(t *testing.T) {
	grass := image.NewRGBA(image.Rect(0, 0, 16, 16))
	tree := image.NewRGBA(image.Rect(0, 0, 16, 16)) // see-through around the trunk
	for i := range grass.Pix {
		grass.Pix[i] = 200
	}
	grass.Pix[3] = 0 // a keyed-clear dark texel: still the ground
	tree.Set(8, 8, color.RGBA{90, 50, 20, 255})
	var g [25][50]Square
	g[3][4].Tile = 1
	g[5][6].Tile = 1
	tiles := []image.Image{grass, tree}
	if got := baseFloor(&g, tiles); got != 0 {
		t.Fatalf("base floor %d, want the grass", got)
	}
}
