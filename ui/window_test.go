package ui

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"gobox/gpu"
)

// A border with no pieces (the game's symbols not loaded yet) draws no
// ribbon rather than crashing.
func TestWindowNoPieces(t *testing.T) {
	dst := gpu.NewImage(64, 64)
	if r := Window(dst, Ribbon{}, gpu.GeoM{}, 8, 8, nil); r.Dx() != 48 {
		t.Fatalf("inside %v", r)
	}
}

// A cap replaces the side piece at the side's first and last cell.
func TestWindowCaps(t *testing.T) {
	px := func(c color.RGBA) *gpu.Image {
		m := image.NewRGBA(image.Rect(0, 0, 8, 8))
		draw.Draw(m, m.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
		return gpu.NewImageFromImage(m)
	}
	red, blue, green := color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}, color.RGBA{0, 255, 0, 255}
	var r Ribbon
	r[Left], r[LeftTopCap], r[LeftBottomCap] = px(green), px(red), px(blue)
	dst := gpu.NewImage(32, 32)
	Window(dst, r, gpu.GeoM{}, 4, 4, nil)
	for _, c := range []struct {
		y    int
		want color.RGBA
	}{{8, red}, {16, blue}} {
		if got := color.RGBAModel.Convert(dst.At(2, c.y+2)); got != c.want {
			t.Fatalf("cell at y=%d: %v, want %v", c.y, got, c.want)
		}
	}
}
