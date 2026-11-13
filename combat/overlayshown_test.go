package combat

import (
	"image"
	"testing"

	"gobox/gpu"
)

// TestOverlayOverFloor: a ground overlay (a body) shows over the floor
// tiles, not hidden under them (POD's bodies never showed).
func TestOverlayOverFloor(t *testing.T) {
	dst := gpu.NewImage(640, 400)
	defer dst.Deallocate()
	body := image.NewRGBA(image.Rect(0, 0, 24, 24))
	for i := 0; i < len(body.Pix); i += 4 {
		body.Pix[i], body.Pix[i+2], body.Pix[i+3] = 255, 255, 255
	}
	floor := image.NewRGBA(image.Rect(0, 0, tilePx, tilePx))
	for i := 0; i < len(floor.Pix); i += 4 {
		floor.Pix[i+1], floor.Pix[i+3] = 80, 255
	}
	b := &Board{Tiles: []image.Image{floor, body}, Units: []Unit{{ID: 1, X: 10, Y: 10, Size: 1, Party: true, Icon: floor, Name: "A", HP: 9, MaxHP: 9}}}
	for y := range b.Ground {
		for x := range b.Ground[y] {
			b.Ground[y][x].Tile = 0
		}
	}
	for y := 0; y < len(b.Ground); y++ {
		for x := range b.Ground[y] {
			b.Ground[y][x].Overlay = 2
		}
	}
	var p CombatPresenter
	p.Draw(dst, b)
	p.Draw(dst, b)
	px := make([]byte, 4*640*400)
	dst.ReadPixels(px)
	n := 0
	for i := 0; i < len(px); i += 4 {
		if px[i] > 150 && px[i+2] > 150 && px[i+1] < 80 {
			n++
		}
	}
	if n == 0 {
		t.Fatal("no overlay pixel over the floor")
	}
}
